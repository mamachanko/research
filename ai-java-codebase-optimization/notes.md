# Notes: AI Agent Optimization for Large Java Codebases

## Topic
How to compensate for Gemini 3.1 Pro's (and any weaker model's) shortcomings when
working with large Java 17, Maven multi-module, Spring Boot 4 codebases.
Techniques applicable to any agent tooling: Cursor CLI, Claude Code, Codex, etc.

---

## Root Problem Analysis

### Why weaker models fail on large Java codebases

**1. Context exhaustion**
- Multi-module Maven projects have hundreds of files; a weaker model fills its
  context window with irrelevant code before reaching the actually important files.
- Spring Boot auto-configuration generates a massive implicit classpath; the model
  cannot "see" beans that are wired in unless explicitly shown.
- Entity graphs (JPA relationships, cascades) are spread across multiple files —
  the model loses the graph before it can reason about it.

**2. Namespace and API staleness**
- Models trained before 2023 internalized `javax.*`; Spring Boot 3+ requires
  `jakarta.*`. Weaker/smaller models revert to old patterns more often.
- Spring Boot 4 (based on Spring Framework 6.2+) will introduce further changes
  (virtual threads as default, Project Loom integration, new `@HttpExchange` idioms,
  Observability API). Models have sparse training data here.
- Java 17 LTS features (records, sealed classes, pattern matching) are used
  inconsistently — models often generate verbose pre-17 equivalents.

**3. Shallow cross-module reasoning**
- A change to a shared `domain` module has ripple effects in `api`, `service`,
  `persistence` modules. Weaker models don't trace these automatically.
- BOM / dependency-management: the model doesn't know which version of a library
  is actually on the classpath unless told explicitly.

**4. Spring "magic" opacity**
- `@ConditionalOnProperty`, `@Profile`, `@EnableAutoConfiguration` exclusions —
  the effective application context at runtime differs from what source code shows.
  Weaker models assume all beans are always active.
- `@Transactional` boundary analysis requires understanding the proxy model;
  weaker models routinely suggest calling `@Transactional` methods from within the
  same bean (self-invocation problem).

**5. Test blindness**
- Models edit production code and forget to update or add tests.
- Models generate tests that compile but don't actually exercise the changed logic.

---

## Strategy Taxonomy

### Tier 1 — Context Engineering (Highest ROI, zero tooling cost)

#### A. Project Bible file (AGENTS.md / CLAUDE.md / .cursorrules / system prompt)

The single most impactful lever. This file is injected at the start of every
agent session. It should contain:

```markdown
## Project Identity
- Java 17, Spring Boot 4.x, Spring Framework 6.2+
- Namespace: jakarta.* everywhere. NEVER use javax.*
- Build: Maven 3.9+, multi-module
- Module layout:
    parent/
      domain/          ← Entities, value objects, domain services (no Spring deps)
      application/     ← Use cases, ports (interfaces), application services
      infrastructure/  ← Adapters: JPA repos, REST clients, messaging
      api/             ← Spring MVC / WebFlux controllers, DTOs, OpenAPI
      bootstrap/       ← Spring Boot main class, configuration, assembly

## Dependency Rules
- domain has ZERO external dependencies (pure Java)
- application depends only on domain
- infrastructure depends on application + domain
- api depends on application + domain
- NEVER add Spring annotations to domain module

## Key Library Versions (from parent POM BOM)
- spring-boot: 4.0.x
- spring-framework: 6.2.x
- hibernate: 6.4.x (jakarta persistence)
- mapstruct: 1.5.x
- testcontainers: 1.20.x

## Code Conventions
- Use Java records for DTOs and value objects
- Use sealed interfaces for discriminated unions / ADTs
- Pattern match with switch expressions, not instanceof chains
- Prefer constructor injection (no @Autowired on fields)
- @Transactional only on application service layer, never on domain

## Common Pitfalls — Do Not Do These
- Do NOT call a @Transactional method from within the same bean
- Do NOT use javax.* imports
- Do NOT add Spring annotations to domain module classes
- Do NOT use Lombok (project does not use it)
- Do NOT use field injection (@Autowired on fields)

## Test Strategy
- Unit tests: JUnit 5 + Mockito, no Spring context
- Integration tests: @SpringBootTest + Testcontainers
- Always update or add tests when modifying behavior
```

#### B. Module README files
Each Maven module gets a `README.md` with:
- What this module is responsible for
- Key classes and their roles
- What it depends on and what depends on it
- Non-obvious decisions / gotchas

The agent reads these at the start of a task scoped to that module.

#### C. Architecture Decision Records (ADRs)
Put `docs/adr/` in the repo. Compact ADRs (10-20 lines each) explain *why* things
are the way they are. Inject relevant ADRs as context when asking the agent to
change something architectural.

#### D. Inline context injection
Before giving a task, manually paste:
```
# Current dependency graph (domain module)
Order → OrderLine → Product (immutable VO)
Order → Customer (aggregate root)
Order.status: OrderStatus (sealed interface: Pending, Confirmed, Shipped, Cancelled)

# Relevant interfaces
interface OrderRepository { ... }
interface PaymentGateway { ... }
```

This "working memory scaffold" costs 200 tokens and saves thousands of incorrect ones.

---

### Tier 2 — Task Decomposition & Scoping

#### A. One module per session
Never ask the agent to "implement feature X end to end." Instead:
1. Session 1: "Define the domain model changes in `domain` module only."
2. Session 2: "Define the port interfaces in `application` module."
3. Session 3: "Implement the JPA adapter in `infrastructure` module."
4. Session 4: "Wire the controller in `api` module."

Each session starts fresh with the Project Bible + module README.

#### B. Analysis before implementation
Split tasks explicitly:
- **Step 1 (analysis)**: "List all classes that need to change if we add a
  `discountCode` field to `Order`. Do not write any code yet."
- **Step 2 (contract)**: "Define the interface/record changes only."
- **Step 3 (implement)**: "Implement the changes in [specific class]."

Weaker models produce better output when forced into the analysis step first —
they commit to a plan and then execute it, rather than improvising.

#### C. Explicit scope declaration
Every prompt should declare:
```
Scope: only modify files in src/main/java/com/example/order/
Do not touch: OrderController, OrderMapper, any test files
After finishing, list all files you modified.
```

#### D. Contract-first
Define interfaces and records before asking for implementations:
```java
// Define this first, have the model confirm it's correct:
public record CreateOrderCommand(CustomerId customerId, List<OrderLineRequest> lines) {}
public interface CreateOrderUseCase {
    OrderId execute(CreateOrderCommand command);
}
```
Now the model implements against a fixed contract, not a moving target.

---

### Tier 3 — Feedback Loops (Compile-Driven Development)

#### A. Compile loop
After every agent edit, run:
```bash
mvn compile -pl <changed-module> -am --no-transfer-progress 2>&1 | tail -40
```
Feed compiler errors back to the agent immediately. Do NOT accumulate errors across
multiple files before feeding back — fix one compile error at a time.

The agent is much better at fixing a specific error than at predicting all errors
upfront.

#### B. Test loop
```bash
mvn test -pl <module> -Dtest=OrderServiceTest --no-transfer-progress 2>&1 | tail -60
```
Run the specific test for the changed class. Red → feed output back → iterate.

#### C. Checkstyle / PMD as guardrails
Configure `maven-checkstyle-plugin` to run on `validate` phase. The agent learns
quickly from checkstyle errors because they are precise ("line 42: 'javax' import
not allowed").

A `.editorconfig` file also helps maintain consistent formatting that the agent
won't fight.

#### D. ArchUnit for architecture enforcement
```java
@ArchTest
static final ArchRule domainHasNoDependencies =
    classes().that().resideInAPackage("..domain..")
             .should().onlyDependOnClassesThat()
             .resideInAnyPackage("..domain..", "java..");
```
Run ArchUnit tests after agent edits. Feed violations back. This is the fastest
way to catch cross-module dependency violations.

---

### Tier 4 — Retrieval & Code Indexing

#### A. ctags / tree-sitter index
Generate a tags file and give the agent grep-friendly class/method lookup:
```bash
ctags -R --languages=Java --fields=+n src/
grep -n "class OrderService" tags
```
For weaker models, providing the exact file:line reference for key classes
eliminates hallucinated locations.

#### B. Dependency graph injection
Generate the module dependency graph once and keep it in the Project Bible:
```bash
mvn dependency:tree -pl domain --no-transfer-progress
```
ASCII-art module graph (draw it manually or generate it) costs ~100 tokens and
prevents the model from suggesting impossible cross-module dependencies.

#### C. Spring bean inventory
For complex auto-configuration scenarios, generate a bean list:
```bash
# Run with spring.main.lazy-initialization=false and actuator
curl -s localhost:8080/actuator/beans | jq '[.contexts[].beans | keys[]]' | sort
```
Paste the relevant beans as context when debugging DI issues.

#### D. RAG / embeddings (for very large codebases)
Tools like `continue.dev`, Cursor's codebase indexing, or a custom `tree-sitter` +
vector DB setup let the agent pull relevant code snippets on demand rather than
loading everything into context. The key is that the *retrieval quality* determines
agent quality — bad retrieval = bad context = bad output.

For Java specifically, index at the method level (not file level) to keep chunks
semantically coherent.

---

### Tier 5 — Prompt Patterns

#### A. The "rubber duck" pattern
Before asking for implementation, ask the model to explain the existing code:
"Explain what `OrderFulfillmentService.fulfill()` does, step by step."
If the explanation is wrong, correct it before proceeding. This surfaces
misunderstandings before they become bugs.

#### B. Adversarial self-review
After the model produces code, ask:
"Review the code you just wrote. Are there any @Transactional self-invocation
problems? Any javax.* imports? Any missing test coverage? List issues."
Weaker models catch ~60% of their own errors when explicitly asked to look.

#### C. Few-shot examples
Provide 1-2 concrete examples of the pattern you want:
"Here is how we implement a use case in this codebase: [paste CreateCustomerService].
Now implement CreateOrderService following exactly this pattern."

Few-shot beats instruction for weaker models more reliably than for stronger ones.

#### D. Negative constraints
Explicitly stating what NOT to do is as important as what to do:
"Do not use Optional.get() without isPresent() check."
"Do not use new ArrayList<>() where List.of() is sufficient."
"Do not use String concatenation in log statements."

#### E. Output format specification
Ask for structured output to force careful thinking:
```
Respond in this format:
1. Files to modify: [list]
2. Files to create: [list]
3. Interfaces that change signature: [list]
4. Tests to update: [list]
5. Implementation: [code]
```

---

### Tier 6 — Agent Tooling Configuration

#### Cursor CLI
- `.cursorrules` file: put the Project Bible here
- Use `@Codebase` sparingly — it pulls too much; prefer `@file` references
- Use `/reference` to pin specific files before asking a question

#### Claude Code
- `CLAUDE.md` at repo root: Project Bible
- Module-level `CLAUDE.md` files for per-module context
- Use `/init` to let Claude index the repo before first use
- Compact conversations frequently to avoid context drift

#### Codex / OpenAI
- System prompt injection: Project Bible as system message
- Use `file_search` tool with a pre-built vector store of module READMEs
- Temperature 0.1-0.2 for implementation tasks

#### Universal (any tool)
- Never start a new session without re-injecting the Project Bible
- Keep sessions short and focused (one module, one task)
- Use git commits as session boundaries — commit working state before a new task

---

## Spring Boot 4 Specific Gotchas

- **Virtual threads**: Boot 4 enables Project Loom virtual threads by default for
  Tomcat and `@Async`. ThreadLocal-based patterns (MDC, security context) need
  explicit configuration. Tell the agent: "virtual threads are enabled; do not use
  ThreadLocal patterns without ScopedValue or explicit propagation."

- **Observability API**: `Micrometer Observation API` replaces direct Timer/Counter
  usage. Models trained before 2024 will suggest old patterns.

- **`@HttpExchange`**: Spring 6 declarative HTTP client. Models will suggest
  Feign/Retrofit. Explicitly tell them to use `@HttpExchange` + `HttpServiceProxyFactory`.

- **Problem Details (RFC 9457)**: Boot 3.2+ auto-configures `ProblemDetail` error
  responses. Tell the agent not to implement custom error response DTOs.

- **Jakarta persistence 3.2**: `EntityManager` API changes (typed `find()`, etc.).

---

## Summary: Priority Order

1. **Project Bible file** (AGENTS.md/CLAUDE.md/.cursorrules) — do this first, always
2. **Module-scoped sessions** — never span multiple modules in one task
3. **Analysis before implementation** — force the model to plan before coding
4. **Compile loop feedback** — mvn compile errors fed back immediately
5. **Contract-first** — define interfaces/records before implementations
6. **Explicit negative constraints** — tell the model what NOT to do
7. **Few-shot examples** — paste a canonical existing implementation
8. **ArchUnit + Checkstyle** — machine-enforced guardrails
9. **Adversarial self-review** — ask model to critique its own output
10. **ctags / bean inventory** — precise file:line references eliminate hallucinations
