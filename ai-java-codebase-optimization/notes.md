# AI-Java Codebase Optimization: Steering & Steroids for Large Complex Codebases

## Problem Statement
- Gemini 3.1 Pro (and other LLMs) struggle with:
  - Multi-module Maven projects with hundreds of files
  - Deep Spring Boot framework patterns
  - Complex dependency graphs
  - Context window limitations
  - Task hallucination on unfamiliar patterns

## Universal Steroids for LLM-Assisted Development

### 1. CODEBASE DECOMPOSITION & SCOPING

**Problem**: Dumping entire codebase info creates noise, confusion, hallucinations.

**Solution**:
- **Map the architecture first**: Before asking AI, create a blueprint
  - Module dependency graph (using `mvn dependency:tree`)
  - Layer structure (controllers → services → repositories → entities)
  - Key entry points and flows
  - Domain boundaries
  
- **Compartmentalize requests**: One module/feature at a time
  - Instead of: "optimize the entire codebase"
  - Do: "optimize the payment-service module's transaction handling"
  
- **Use architectural documentation**:
  - Create architecture.md showing module relationships
  - Document layer responsibilities
  - Pass this to AI before deep dives

**Applicable to**: All tools (Cursor, Claude Code, Codex, etc.)

---

### 2. SMART CONTEXT SEEDING

**Problem**: LLMs waste tokens re-discovering obvious patterns.

**Solution**:
- **Preflight grep/glob operations**: Gather relevant code before asking AI
  - Find all Spring `@Service` classes: `grep -r "@Service" src/`
  - Find configuration classes: `grep -r "@Configuration" src/`
  - Find persistence patterns: `grep -r "extends JpaRepository" src/`
  - Pass curated results to AI instead of full folder dumps

- **Create context documents**:
  ```
  ## Naming Conventions
  - Services: *Service.java
  - DAOs: *Repository.java
  - DTOs: *Request/Response/DTO.java
  - Configs: *Config.java
  
  ## Key Patterns Used
  - Spring Data JPA for persistence
  - @Aspect for cross-cutting concerns
  - RestTemplate for HTTP (legacy), WebClient for reactive
  - Maven profiles: dev, staging, prod
  ```

- **Provide type maps**:
  ```
  Key Classes:
  - User → com.company.user.entity.User
  - Payment → com.company.payment.entity.Payment
  - AuditLog → com.company.audit.AuditLog
  ```

**Applicable to**: All tools

---

### 3. EXPLICIT BOUNDARIES & CONSTRAINTS

**Problem**: AI explores tangential optimizations, suggests inconsistent patterns.

**Solution**:
- **Define scope explicitly**:
  ```
  SCOPE:
  - Only touch: payment-service/src/main/java
  - DO NOT modify: shared-lib, common-entities
  - Constraint: Must maintain Spring's current version (5.3.x)
  - Constraint: Cannot add new dependencies
  ```

- **State non-negotiables**:
  ```
  FIXED PATTERNS (do not suggest refactoring):
  - We use JpaRepository, not custom SQL
  - We use @Transactional extensively
  - We enforce request/response DTOs
  - We use Aspect-based logging (don't suggest decorators)
  ```

- **Performance constraints**:
  ```
  Must maintain:
  - Response time < 200ms for user endpoints
  - Database queries: max 3 per request
  - Memory footprint: < 512MB heap
  ```

**Applicable to**: All tools

---

### 4. ARCHITECTURAL PATTERNS DOCUMENTATION

**Problem**: AI reinvents patterns already established in the codebase.

**Solution**:
- **Create a PATTERNS.md file**:
  ```markdown
  ## Service Layer Pattern
  @Service
  public class PaymentService {
    - Always use constructor injection
    - One repository per domain entity
    - Business logic here, not in controller
    - @Transactional on write operations
    - Separate query methods with @Transactional(readOnly=true)
  }
  
  ## Error Handling Pattern
  - Custom exceptions: PaymentException, ValidationException
  - Never throw raw SQLExceptions
  - Always log with MDC context
  - Return ApiError DTOs with specific error codes
  
  ## Testing Pattern
  - Unit tests: mock all dependencies (@ExtendWith(MockitoExtension.class))
  - Integration tests: use @SpringBootTest with testcontainers
  - Never use H2 for testing, use Testcontainers for real DB
  ```

- **Pass this to AI before asking for new code**

**Applicable to**: All tools

---

### 5. INCREMENTAL PROBLEM BREAKDOWN

**Problem**: LLMs fail at complex multi-step tasks, producing partial solutions.

**Solution**:
- **Break into digestible chunks**:
  - WRONG: "Refactor the entire Order domain for event sourcing"
  - RIGHT: 
    1. Identify all Order-related queries and aggregates
    2. Design OrderEvent hierarchy
    3. Implement EventStore interface
    4. Migrate one feature (e.g., order creation) to event-driven
    5. Add event replay mechanism
    6. Integrate with existing code

- **Checkpoint after each step**: Review AI output before proceeding

- **Use conversation history strategically**: 
  - Summarize decisions at each checkpoint
  - Reference earlier decisions to maintain consistency

**Applicable to**: All tools (critical for Cursor CLI, Claude Code)

---

### 6. FORCED EXPLORATION & DISCOVERY

**Problem**: AI guesses without exploring actual codebase structure.

**Solution**:
- **Make exploration a first step**:
  ```
  TASK 1: Exploration (don't code yet)
  1. List all Spring Configuration classes and what they configure
  2. Show all @Aspect classes and what they intercept
  3. Show all custom annotations and their use sites
  4. List all non-standard dependencies and their purpose
  ```

- **Use glob/grep output as "context oracle"**:
  - Before proposing a solution, AI should reference actual grep results
  - "I found 12 custom exception classes in src/main/java/com/company/exception/"
  - "Pattern: All Services extend BaseService with logging support"

- **Demand specificity**:
  - Don't accept: "You could use a repository pattern here"
  - Demand: "Show me how this fits with the existing UserRepository and PaymentRepository classes"

**Applicable to**: Claude Code, Cursor CLI (with tool access)

---

### 7. MAVEN-SPECIFIC STEROID

**Problem**: Multi-module Maven projects cause confusion; AI doesn't understand module boundaries.

**Solution**:
- **Create a pom-map.txt**:
  ```
  pom.xml (root)
    └── payment-service/ (packaging: jar, depends: common-lib)
         └── pom.xml (uses spring-boot-maven-plugin)
    └── user-service/ (packaging: jar, depends: common-lib)
    └── common-lib/ (packaging: jar, no spring)
    └── api-contracts/ (packaging: jar, DTOs only)
  ```

- **Tell AI about build behavior**:
  ```
  BUILD FACTS:
  - `mvn clean install` builds all modules in dependency order
  - Tests run with `mvn test`
  - Integration tests marked @Tag("integration") run separately
  - Docker build happens in CI/CD, not locally
  ```

- **Explicitly state module access rules**:
  ```
  IMPORT RULES:
  - payment-service can import: common-lib, api-contracts
  - user-service can import: common-lib, api-contracts
  - common-lib imports: nothing from other modules
  - This prevents circular dependencies
  ```

**Applicable to**: All tools

---

### 8. SPRING BOOT SPECIFIC STEROID

**Problem**: Spring magic obscures what's really happening; AI generates incorrect assumptions.

**Solution**:
- **Provide Spring facts**:
  ```
  SPRING VERSION: 5.3.x (LTS)
  BOOT VERSION: 2.7.x
  AUTO-CONFIG: Enabled
  ACTUATOR: Enabled on port 9090
  
  Key Beans (created automatically):
  - RestTemplate (from @Bean in config)
  - DataSource (auto-config from DB properties)
  - JpaTransactionManager (auto-config)
  - RequestMappingHandlerMapping (auto)
  
  NOT auto-configured (we provide):
  - Custom RequestInterceptor
  - PaymentServiceImpl
  - AuditAspect
  ```

- **Document configuration sources**:
  ```
  application.properties: core settings
  application-dev.properties: dev overrides
  application-prod.properties: production overrides
  @Configuration classes: programmatic beans
  ```

- **Explain property binding**:
  ```
  payment.max-retries=3 → injected via @ConfigurationProperties("payment")
  spring.datasource.url → DataSource configuration
  ```

**Applicable to**: All tools

---

### 9. TEST-DRIVEN VALIDATION

**Problem**: AI code passes code review but fails real usage.

**Solution**:
- **Before implementation, write test shapes**:
  ```java
  @Test
  void testPaymentProcessing_WithValidAmount_Succeeds() {
    // Structure defined, AI implements body
  }
  
  @Test
  void testPaymentProcessing_WithNullReference_ThrowsException() {
    // Structure defined, AI implements body
  }
  ```

- **Make AI write tests first** (TDD approach):
  - Ask: "Write tests for PaymentService.processPayment(), then implement it"
  - Tests clarify intent better than requirements
  - Forces AI to understand the domain

- **Validate with actual runs**:
  - Don't just review code, run it
  - `mvn test` catches hallucinated APIs quickly
  - Integration tests catch Spring-specific issues

**Applicable to**: Claude Code, Cursor CLI

---

### 10. DOMAIN MODEL CLARITY

**Problem**: AI generates anemic models or violates DDD principles.

**Solution**:
- **Document your domain model upfront**:
  ```markdown
  ## Payment Domain
  
  **Aggregates**:
  - Payment (root): contains PaymentStatus state machine
    - Cannot transition InvalidState → Completed
    - Must emit PaymentProcessedEvent when completed
    - Validates amounts > 0
  
  **Value Objects**:
  - Money: amount + currency, immutable
  - PaymentStatus: enum, with valid transitions
  
  **Repositories**:
  - PaymentRepository: find by ID, by User, by Status
  
  **Services**:
  - PaymentProcessor: orchestrates payment flow
  - PaymentValidator: validates amounts and rules
  ```

- **Pass this BEFORE asking for code**

- **Reference domain language in requests**:
  - "Implement Payment.transitionTo(PaymentStatus.COMPLETED)"
  - Not: "Update the status field"

**Applicable to**: All tools

---

## Tool-Specific Optimizations

### Claude Code Specific
- Use `/find` to explore structure, reference results in prompts
- Use file watchers to catch errors immediately
- Commit frequently; use git history to rollback hallucinations
- Ask Claude to "explain the module layout first" before coding

### Cursor CLI Specific
- Use `cursor.rules` file to lock in patterns
- Use `cursor find` output as context before asking changes
- Use "generate tests" feature to validate implementations
- Run linters continuously (`eslint --watch` equivalent for Java: `mvn spotbugs:gui`)

### Codex/API-Based Tools
- Prepare prompt templates with fixed preambles (patterns, constraints)
- Use few-shot examples: "Here's how we implement Services, now implement X"
- Keep system prompts under 2KB; use files for context
- Temperature: 0.2 for consistency, 0.5 for exploration

---

## Red Flags & Anti-Patterns

### What Fails:
1. ❌ Dumping entire src/ folder as context
2. ❌ Asking "how would you refactor this?" without constraints
3. ❌ Expecting AI to understand cyclomatic complexity issues
4. ❌ Asking for best practices without defining your standards
5. ❌ Mixing Java 8, 11, 17 patterns without clarity
6. ❌ Not running tests after AI code generation
7. ❌ Letting AI modify core infrastructure without review

### What Works:
1. ✅ Narrow, specific tasks with clear boundaries
2. ✅ Pre-exploration and context seeding
3. ✅ Explicit constraints and non-negotiables
4. ✅ Test-first approach
5. ✅ Incremental changes with checkpoint reviews
6. ✅ Domain model documentation
7. ✅ Running full test suite after every change

---

## Applied Workflow Example

**Task**: Add audit logging to PaymentService

1. **EXPLORATION** (AI)
   - Find all @Service classes; show PaymentService structure
   - Find all @Aspect classes; show logging implementations
   - Find all audit-related classes

2. **CONTEXT** (Human)
   - "Here's our AuditAspect pattern" [paste code]
   - "Our @Auditable annotation target is METHOD"
   - "AuditLog structure: [user_id, action, resource, timestamp]"

3. **IMPLEMENTATION** (AI, with boundaries)
   - "Add @Auditable to PaymentService methods that change state"
   - "Don't modify payment processing logic"
   - "Must maintain current response time SLA"

4. **VALIDATION** (Human/AI)
   - Run: `mvn clean test`
   - Check audit logs in test output
   - Verify no performance regression

5. **REVIEW** (Human)
   - Check diff; ensure no scope creep
   - Commit with clear message

---

## Summary Table: Problem → Solution → Applicable Tools

| Problem | Solution | Tools |
|---------|----------|-------|
| Context confusion | Pre-explore, provide architecture.md | All |
| Hallucinated patterns | Document established patterns explicitly | All |
| Scope creep | Define fixed boundaries and constraints | All |
| Wrong decisions | Checkpoint after each logical step | All |
| Implementation bugs | Test-first, run mvn test after | Claude Code, Cursor |
| Module confusion | Create pom-map.txt, document rules | All |
| Spring magic issues | Document auto-config and @Beans | All |
| Domain violations | Document DDD aggregates/rules upfront | All |
| Performance regression | Include performance constraints | All |
| Bad abstractions | Force exploration phase before coding | All |
