# Steering AI Agents on Large Java Codebases

How to compensate for weak model performance (e.g. Gemini 3.1 Pro) when working
with Java 17 / Maven multi-module / Spring Boot 4 codebases, and how to get
top-tier outcomes from any agent tooling — Cursor CLI, Claude Code, Codex, or others.

---

## Why Weaker Models Fail on Large Java Codebases

Before reaching for solutions, understand the failure modes:

| Failure Mode | Root Cause |
|---|---|
| `javax.*` imports | Model trained on pre-Boot-3 code; reverts under uncertainty |
| Hallucinated class names | Context exhaustion; model fills gaps with plausible fiction |
| Cross-module dependency violations | Cannot hold the full module graph in working memory |
| `@Transactional` self-invocation bugs | Shallow proxy model understanding |
| Stale Spring patterns | Sparse training data on Boot 3.2+/4.x and Framework 6.2+ |
| Missing test updates | Task framing didn't include tests in scope |
| Wrong BOM version used | No knowledge of which library version the BOM pins |

The core insight: **a weaker model with perfect context outperforms a stronger model
with poor context.** Every strategy below is a form of context surgery.

---

## The Playbook (Prioritized)

### 1. The Project Bible — Highest ROI, Do This First

Create a single file at the repo root that every agent tool will auto-inject as
context. Name it `AGENTS.md`, `CLAUDE.md`, `.cursorrules`, or whatever your tool
uses. This is your most powerful lever.

**Minimum viable Project Bible for a Java Spring project:**

```markdown
## Stack
- Java 17, Spring Boot 4.x, Spring Framework 6.2+
- jakarta.* namespace EVERYWHERE. NEVER use javax.*
- Maven 3.9+ multi-module build
- Virtual threads enabled (Project Loom). Do NOT use ThreadLocal patterns.

## Module Layout
parent/
  domain/          ← Pure Java. Entities, VOs, domain services. ZERO Spring deps.
  application/     ← Use cases, port interfaces (no implementations here)
  infrastructure/  ← JPA, REST clients, messaging adapters
  api/             ← Spring MVC controllers, DTOs, OpenAPI
  bootstrap/       ← @SpringBootApplication, configuration assembly

## Dependency Rules (ArchUnit enforced)
domain ← application ← infrastructure
domain ← application ← api
domain and application MUST NOT depend on Spring Framework classes

## Key BOM Versions
spring-boot: 4.0.x | hibernate: 6.4.x | testcontainers: 1.20.x | mapstruct: 1.5.x

## Conventions
- Java records for DTOs and value objects
- Sealed interfaces for discriminated unions (OrderStatus, etc.)
- Switch expressions with pattern matching, not instanceof chains
- Constructor injection only. NEVER @Autowired on fields.
- @Transactional at application service layer only, never on domain classes
- @HttpExchange for declarative HTTP clients, NOT Feign or Retrofit

## Hard Prohibitions
- NO javax.* imports
- NO field injection (@Autowired on fields)
- NO @Transactional self-invocation (calling @Tx method from same bean)
- NO Spring annotations in domain module
- NO Lombok (not in this project)
- NO Optional.get() without isPresent() / orElseThrow()
```

**Why each section matters:**
- *Stack declaration* forces the model to activate its Spring Boot 4 / Java 17 knowledge
  rather than defaulting to the more common Boot 2.x patterns in its training data.
- *Module layout* prevents cross-layer contamination errors — the single most common
  structural mistake on multi-module projects.
- *Hard Prohibitions* are explicitly more effective than positive instructions for
  weaker models; "never do X" outperforms "prefer Y."

---

### 2. Module-Scoped Sessions — Scope Limits Failure Radius

**Never** ask an agent to implement an end-to-end feature in one session. Slice by
module boundary:

```
Session 1 → domain module only (entities, value objects, domain events)
Session 2 → application module only (port interfaces, use case classes)
Session 3 → infrastructure module only (JPA adapters, repository impls)
Session 4 → api module only (controllers, DTOs, mappers)
```

Each session should begin with the Project Bible + the relevant module's `README.md`.
This keeps context small and relevant. A 500-token context window of perfectly relevant
code beats a 32k window of noise.

**Module README template:**

```markdown
# [module-name] module

**Responsibility**: [one sentence]

**Key classes**:
- `OrderService`: orchestrates order lifecycle, owns @Transactional boundary
- `OrderRepository`: port interface (defined here, implemented in infrastructure)

**Depends on**: domain
**Used by**: api, infrastructure

**Gotchas**:
- OrderService.fulfill() is the only entry point for fulfillment — don't bypass it
- Events are published via DomainEventPublisher, not Spring's ApplicationEventPublisher
```

---

### 3. Analysis Before Implementation — Force a Plan

Weaker models produce substantially better code when forced to reason before acting.
Split every non-trivial task into two prompts:

**Prompt 1 (analysis only):**
```
Given the attached [domain model + current OrderService], list every class and
interface that needs to change if we add a `discountCode: String` field to Order.
Show impact across all modules. Do NOT write any code yet.
```

**Prompt 2 (implement with plan confirmed):**
```
Based on the analysis above, implement the domain module changes only:
- Add discountCode to the Order entity
- Add it to the CreateOrderCommand record
- Update the CreateOrderUseCase interface signature
Do not touch anything in infrastructure or api modules yet.
```

This pattern works because the model commits to a consistent understanding before
generating code, rather than discovering inconsistencies mid-generation and silently
backpedaling.

---

### 4. Contract-First — Pin the API Before the Implementation

Define interfaces and records before asking for implementations. The model fills in
implementations against a fixed target rather than improvising both simultaneously.

```java
// Establish this contract first — paste it into the prompt, ask model to confirm
public record CreateOrderCommand(
    CustomerId customerId,
    List<OrderLineRequest> lines,
    Optional<String> discountCode
) {}

public interface CreateOrderUseCase {
    OrderId execute(CreateOrderCommand command);
}
```

Once confirmed correct, ask for `CreateOrderService implements CreateOrderUseCase`.
The model's degrees of freedom are now bounded; it cannot hallucinate method signatures.

---

### 5. Compile Loop — Fastest Feedback Cycle

Run after every agent edit. Do not accumulate errors:

```bash
# Compile only the changed module and its dependencies
mvn compile -pl application -am --no-transfer-progress 2>&1 | tail -30
```

Feed compiler output back immediately. A specific compiler error ("cannot find symbol:
class javax.persistence.Entity — did you mean jakarta.persistence.Entity?") is
enormously more actionable than "the code doesn't work."

The model fixes one compile error far more reliably than it predicts all errors upfront.

**For tests:**
```bash
mvn test -pl application -Dtest=CreateOrderServiceTest --no-transfer-progress 2>&1 | tail -50
```

---

### 6. Machine-Enforced Guardrails — ArchUnit + Checkstyle

Don't rely on the model to self-police architectural rules. Encode them as tests.

**ArchUnit (add to `architecture-test` module or inline):**
```java
@ArchTest
static final ArchRule domainIsPure =
    classes().that().resideInAPackage("..domain..")
             .should().onlyDependOnClassesThat()
             .resideInAnyPackage("..domain..", "java..", "org.slf4j..");

@ArchTest
static final ArchRule noJavaxImports =
    noClasses().should().dependOnClassesThat()
               .resideInAPackage("javax..");

@ArchTest
static final ArchRule onlyApplicationServicesAreTransactional =
    classes().that().areAnnotatedWith(Transactional.class)
             .should().resideInAPackage("..application.service..");
```

**Checkstyle rule (illegal-imports.xml):**
```xml
<module name="IllegalImport">
    <property name="illegalPkgs" value="javax.persistence, javax.servlet, javax.validation"/>
</module>
```

Run these on every compile. Feed violations back to the model. Violations are machine-
precise and the model handles them well.

---

### 7. Few-Shot Examples — Show, Don't Just Tell

For weaker models, a concrete example is worth five paragraphs of instructions. Before
asking for a new use case, paste a canonical existing one:

```
Here is how CreateCustomerService is implemented in this codebase:
[paste 30-40 lines of the class]

Now implement CreateOrderService following exactly this same pattern:
- Same constructor injection style
- Same @Transactional placement
- Same domain event publishing pattern
- Same exception handling approach
```

This is the single most reliable technique for stylistic consistency across a session.

---

### 8. Adversarial Self-Review — Ask the Model to Critique Itself

After the model produces code, immediately ask:

```
Review the code you just wrote against these criteria:
1. Any javax.* imports? (must be jakarta.*)
2. Any @Transactional self-invocation? (method calling another @Tx method in same bean)
3. Any field injection? (must be constructor injection)
4. Did you update the tests?
5. Any Spring annotations in the domain module?

List any violations found. If none, say "no violations."
```

Weaker models catch approximately 60-70% of their own errors when explicitly asked to
look for specific things. The adversarial framing ("violations found") is more effective
than "is everything correct?"

---

### 9. Inline Context Injection — Cheap Tokens, High Signal

Before any complex task, paste a compact summary of the relevant domain state:

```
# Domain model snapshot (relevant to this task)
Order (aggregate root)
  - orderId: OrderId (record)
  - status: OrderStatus (sealed interface: Pending | Confirmed | Shipped | Cancelled)
  - lines: List<OrderLine> (entity, belongs to Order)
  - customerId: CustomerId (value object, not a Customer reference)

# Key constraint
Order.confirm() may only be called when status == Pending.
Order publishes OrderConfirmed (domain event) on successful confirm().

# Relevant interfaces
OrderRepository.save(Order): void
OrderRepository.findById(OrderId): Optional<Order>
DomainEventPublisher.publish(DomainEvent): void
```

200 tokens of domain model snapshot eliminates 90% of hallucinated class names and
wrong relationship assumptions.

---

### 10. Explicit Output Format — Force Structured Thinking

Ask for structured output on complex tasks:

```
Respond in exactly this format before writing any code:

FILES TO MODIFY:
- [list with reason]

FILES TO CREATE:
- [list with reason]

INTERFACES WITH CHANGED SIGNATURES:
- [list — these affect other modules]

TESTS TO UPDATE:
- [list]

IMPLEMENTATION:
[code below]
```

This forces the model to do impact analysis before it starts generating, catching
cross-module effects that it would otherwise miss.

---

## Spring Boot 4 Specific Injection Points

Add these to your Project Bible if on Boot 4:

```markdown
## Boot 4 / Framework 6.2+ Specifics

### Virtual Threads (enabled by default)
- Tomcat uses virtual threads for request handling
- @Async uses virtual threads
- DO NOT use ThreadLocal for per-request state; use ScopedValue or pass explicitly
- MDC propagation needs explicit InheritableThreadLocal or Micrometer Observation Context

### Declarative HTTP Clients
- Use @HttpExchange + HttpServiceProxyFactory
- NOT Feign, NOT Retrofit, NOT RestTemplate (deprecated)
- RestClient is the replacement for RestTemplate in imperative code

### Observability
- Use Micrometer Observation API (ObservationRegistry)
- NOT micrometer Timer/Counter/Gauge directly in business code
- @Observed annotation for automatic span/metric creation

### Error Handling
- ProblemDetail (RFC 9457) is auto-configured
- DO NOT create custom error response DTOs
- Use @ExceptionHandler returning ProblemDetail or ResponseEntity<ProblemDetail>

### Jakarta Persistence 3.2
- EntityManager.find() is now typed: em.find(Order.class, id) — still the same
- CriteriaBuilder has new type-safe APIs; prefer them over string-based

### Testing
- @MockitoBean replaces @MockBean (Spring Boot 4 deprecates the Mockito integration)
- Use @TestcontainersConfiguration for shared container lifecycle
```

---

## Tooling-Specific Configuration

### Claude Code
```
# CLAUDE.md at repo root → auto-injected every session
# Add module-level CLAUDE.md in each Maven module directory
# Use /compact frequently to prevent context drift on long sessions
# Session discipline: one module per conversation
```

### Cursor CLI
```
# .cursorrules at repo root → injected as system context
# Use @file references to pin specific files (beats @Codebase for focused tasks)
# /reference command to attach module README before each task
```

### Codex / OpenAI Assistants
```
# System message: paste Project Bible
# file_search tool: vector store containing module READMEs + ADRs
# temperature: 0.1 for implementation, 0.3 for analysis
# Use structured output mode for analysis tasks
```

### Any Tool — Universal Session Ritual
```
1. Paste Project Bible (or confirm it's in system context)
2. Paste module README for the module you're working in
3. Paste domain model snapshot relevant to the task
4. State the task with explicit scope and prohibitions
5. Ask for analysis first (no code), confirm, then ask for implementation
6. Run mvn compile, feed errors back
7. Run relevant tests, feed results back
8. Ask for adversarial self-review
9. Commit. Start next session fresh.
```

---

## Quick Reference Card

| Problem | Fix |
|---|---|
| `javax.*` imports appearing | Add `NEVER use javax.*` to Project Bible; add Checkstyle rule |
| Wrong module gets modified | Declare scope explicitly: "only files in application/src/main/..." |
| Hallucinated class names | Paste ctags output or file:line references for key classes |
| `@Transactional` bugs | ArchUnit rule + add to adversarial review checklist |
| Missing tests | Add "update relevant tests" to every task prompt |
| Wrong library version | Paste `mvn dependency:tree` output for the module |
| Spring bean not found | Paste `actuator/beans` output; describe auto-configuration exclusions |
| Stale Spring Boot 2 patterns | Few-shot example from existing Boot 4 class in repo |
| Cross-module ripple missed | Use structured output format; explicitly ask for "interfaces that change" |
| Model loses context mid-session | Commit, start new session, re-inject Project Bible |

---

## Key Insight

The gap between a weak model and a strong model is **smaller than the gap between a
weak model with good context and a weak model with poor context.** Gemini 3.1 Pro with
a well-crafted Project Bible, module-scoped sessions, contract-first prompting, and a
compile feedback loop will outperform Claude Sonnet or GPT-4o operating blind in a
sprawling repo with no guardrails.

The techniques above are universally applicable and compound: each one independently
improves outcomes, and together they create a system where even a mediocre model is
constrained enough to produce production-quality Java code.
