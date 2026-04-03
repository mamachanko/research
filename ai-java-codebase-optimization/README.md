# AI-Java Codebase Optimization: Complete Guide

## Executive Summary

Working with large Java/Spring Boot codebases (17+, Maven multi-module, 100s-1000s of files) using LLM-based tools (Gemini 3.1 Pro, Claude, Cursor, Codex) requires **deliberate strategies** to overcome inherent limitations. Rather than accepting mediocre results, this research identifies 10 universal "steroids" (techniques) applicable across all tooling that dramatically improve outcomes.

**Key Finding**: The quality gap is not inherent to the LLM—it's the **input quality and framing**. Poor results come from poor prompting and missing context, not tool limitations.

---

## The Core Problem

### What Fails with Large Codebases

1. **Context Collapse**: Dumping entire module directory → AI generates noise, hallucinations, wrong patterns
2. **Pattern Blindness**: AI re-invents patterns already in the codebase instead of following them
3. **Scope Creep**: Without explicit constraints, AI optimizes tangentially or suggests unnecessary refactoring
4. **Architectural Ignorance**: AI doesn't know module boundaries, dependency rules, or layer responsibilities
5. **Spring Magic**: Auto-config, component scanning, bean creation happen invisibly; AI makes wrong assumptions
6. **No Grounding**: AI guesses structure instead of exploring actual code
7. **Test Blindness**: Code looks good but breaks at runtime due to Spring wiring, JPA behavior, etc.

### Why Gemini 3.1 Pro Struggles (and so do others)

- No inherent understanding of your codebase patterns
- Context windows limit how much code can be shown
- Without guidance, explores low-probability paths
- Cannot distinguish between "best practices" and "your standards"
- Lacks ability to validate code without test execution

---

## The 10 Universal Steroids

### 1. **Codebase Decomposition & Scoping**
   - **Problem**: Entire-codebase context = noise
   - **Solution**: Map architecture first, then ask about one module at a time
   - **Result**: AI focuses effort, fewer hallucinations
   - **Tool Agnostic**: Yes

### 2. **Smart Context Seeding**
   - **Problem**: AI wastes tokens re-discovering obvious patterns
   - **Solution**: Pre-grep patterns, provide curated context documents
   - **Result**: AI operates on "known facts" not guesses
   - **Example**: Before asking for code, provide list of all @Service classes found via grep

### 3. **Explicit Boundaries & Constraints**
   - **Problem**: AI explores tangential optimizations
   - **Solution**: Define scope, fixed patterns, performance constraints explicitly
   - **Result**: AI stays in lanes, no scope creep
   - **Template**: "SCOPE: [modules to touch] | CONSTRAINTS: [what not to change] | PERFORMANCE: [SLAs]"

### 4. **Architectural Patterns Documentation**
   - **Problem**: AI reinvents patterns instead of following them
   - **Solution**: Create PATTERNS.md documenting your established approaches
   - **Result**: AI extends patterns, doesn't create new ones
   - **Content**: Service layer pattern, error handling, testing, transaction boundaries

### 5. **Incremental Problem Breakdown**
   - **Problem**: LLMs fail at complex multi-step tasks
   - **Solution**: Break into digestible chunks, checkpoint between steps
   - **Result**: Each step validated, consistent decisions maintained
   - **Practice**: Complex task = 5-7 small subtasks, each reviewed before next

### 6. **Forced Exploration & Discovery**
   - **Problem**: AI guesses without exploring actual structure
   - **Solution**: Make exploration phase mandatory before implementation
   - **Result**: AI references real code, not assumptions
   - **Pattern**: "TASK 1: Explore [list what to find] | TASK 2: Implement based on findings"

### 7. **Maven-Specific Steroid**
   - **Problem**: Multi-module projects cause confusion about boundaries
   - **Solution**: Provide pom-map.txt showing module hierarchy and import rules
   - **Result**: AI understands which modules can import which
   - **Bonus**: Document build facts (profiles, test tags, CI behavior)

### 8. **Spring Boot Specific Steroid**
   - **Problem**: Spring magic (auto-config, component scanning) invisible to AI
   - **Solution**: Document Spring facts, beans, configuration sources, property binding
   - **Result**: AI doesn't hallucinate about bean creation or wiring
   - **Content**: Versions, auto-config settings, key beans, property sources

### 9. **Test-Driven Validation**
   - **Problem**: Code reviews pass; runtime fails
   - **Solution**: Write test shapes first, make AI implement tests before code, run mvn test
   - **Result**: Bugs caught immediately, Spring behavior validated
   - **Practice**: "Write tests for X, then implement X" forces understanding

### 10. **Domain Model Clarity**
   - **Problem**: AI generates anemic models or violates DDD principles
   - **Solution**: Document aggregates, value objects, invariants, transitions upfront
   - **Result**: AI generates semantically correct domain logic
   - **Content**: Use DDD language (aggregates, repositories, domain events)

---

## Applied Workflow: End-to-End Example

### Task: Add Audit Logging to PaymentService

#### Phase 1: Preparation (5 min)
```
Create AUDIT-CONTEXT.md:

## Audit Context

### Current State
- AuditAspect exists at com.company.audit.AuditAspect
- Uses @Auditable annotation, targets METHOD
- AuditLog entity has: userId, action, resourceType, resourceId, timestamp, details

### Constraints
- Do NOT modify PaymentService business logic
- Do NOT add new database tables
- Must maintain response time < 200ms
- Do NOT change existing transaction boundaries

### Where to Add
- PaymentService methods that mutate state:
  - processPayment()
  - refundPayment()
  - updatePaymentStatus()
- Do NOT log query methods
```

#### Phase 2: Exploration (AI, 5 min)
```
Prompt: "Using the AUDIT-CONTEXT.md provided:
1. Find all @Service classes and list 5 of them
2. Show the @Auditable pattern used in any existing service
3. Show the AuditLog entity structure
4. List PaymentService's current public methods

Reference actual code paths, not assumptions."
```

#### Phase 3: Implementation (AI, 10 min)
```
Prompt: "Based on exploration results:
1. Add @Auditable annotations to PaymentService.processPayment(), 
   refundPayment(), updatePaymentStatus()
2. Use action names: 'PAYMENT_PROCESSED', 'PAYMENT_REFUNDED', 'PAYMENT_STATUS_UPDATED'
3. Do NOT modify method signatures or business logic
4. Ensure annotations are valid (check syntax against existing @Auditable examples)"
```

#### Phase 4: Validation (Human + AI, 10 min)
```bash
# Run tests
mvn clean test -Dtest=PaymentServiceTest

# Check audit logs in test output
grep -A2 "PAYMENT_PROCESSED" target/test-logs.txt

# Verify no performance regression
# Check before/after: mvn clean test -DargLine="-javaagent:jprofileragent.jar"
```

#### Phase 5: Review (Human, 5 min)
```
git diff src/main/java/com/company/payment/PaymentService.java
# Should see: only @Auditable added, no logic changes
# Commit with: "Add audit logging to PaymentService mutations"
```

**Total time**: ~35 min vs. 2-3 hours of back-and-forth with unfocused AI

---

## Template Files for Your Codebase

### 1. ARCHITECTURE.md Template
```markdown
# Architecture Overview

## Module Structure
- **api-contracts**: DTOs and interfaces, no logic
- **common-lib**: Shared utilities, no Spring dependencies
- **payment-service**: Payment domain, depends on common-lib + api-contracts
- **user-service**: User domain, depends on common-lib + api-contracts

## Layer Structure
```
Controller → Service → Repository → Entity
         ↓
      Aspect (logging, security)
```

## Dependencies
```
payment-service can import:
  ✓ com.company.common.*
  ✓ com.company.api.*
  ✗ com.company.user.*
```

## Key Patterns
- Immutable DTOs
- Spring Data JPA for persistence
- Aspect-based cross-cutting concerns
- Custom exceptions for error handling
```

### 2. PATTERNS.md Template
```markdown
# Established Patterns

## Service Layer
- Constructor injection only
- @Transactional on write operations
- @Transactional(readOnly=true) on queries
- No direct entity return; use DTOs

## Error Handling
- Custom exceptions: PaymentException, ValidationException
- Log at service layer, rethrow with context
- GlobalExceptionHandler returns ApiError DTO

## Testing
- Unit: @ExtendWith(MockitoExtension.class), mock dependencies
- Integration: @SpringBootTest with Testcontainers
- Never use H2 in tests; use real DB via containers
```

### 3. MAVEN-FACTS.md Template
```markdown
# Maven Build Facts

## Multi-Module Structure
```
pom.xml (root, packaging: pom)
├── payment-service/ (packaging: jar)
├── user-service/ (packaging: jar)
├── common-lib/ (packaging: jar)
└── api-contracts/ (packaging: jar)
```

## Build Profiles
- dev: local development, debug logging
- staging: pre-production, info logging
- prod: production, warn+ logging, optimizations

## Module Import Rules
```
payment-service → common-lib, api-contracts ✓
user-service → common-lib, api-contracts ✓
common-lib → (nothing) ✓
api-contracts → (nothing) ✓
```

## Build Commands
- `mvn clean install`: full build
- `mvn test`: all unit tests
- `mvn verify -DskipITs=false`: include integration tests
- `mvn -pl payment-service clean test`: single module
```

### 4. SPRING-FACTS.md Template
```markdown
# Spring Configuration Facts

## Versions
- Spring: 5.3.x (LTS)
- Boot: 2.7.x
- Java: 17

## Auto-Config
- Enabled: DataSource, JPA, Web, Actuator
- Disabled: JMX, Security (custom config)

## Key Beans (Auto-Created)
- DataSource from properties
- JpaTransactionManager
- RestTemplate
- RequestMappingHandlerMapping

## Custom Beans (from @Configuration)
- PaymentServiceImpl
- AuditAspect
- GlobalExceptionHandler

## Property Binding
- payment.max-retries (via @ConfigurationProperties)
- spring.datasource.url (DataSource)
- logging.level.* (Logback)
```

---

## Tool-Specific Implementation Strategies

### Using Claude Code
```
1. Start with `/find src/main/java` to explore structure
2. Reference find results in all subsequent prompts
3. Use git diff frequently to validate changes
4. Commit after logical chunks, not per-file
5. Run tests immediately after: `mvn test`
```

### Using Cursor CLI
```
1. Place ARCHITECTURE.md, PATTERNS.md, MAVEN-FACTS.md in project root
2. Create .cursor/rules (if exists) with pattern documentation
3. Use `cursor find "pattern"` before asking for changes
4. Use "Generate Tests" feature for validation
5. Commit frequently with clear messages
```

### Using API-Based Tools (Codex, etc.)
```
1. Prepare system prompt preamble:
   - Module structure
   - Import rules
   - Naming conventions
   - Key classes

2. Use few-shot examples:
   "Here's how we implement Services:
   [example Service class]
   Now implement PaymentService with: [requirements]"

3. Temperature settings:
   - 0.2 for consistency (refactoring, pattern following)
   - 0.5 for exploration (new features, alternatives)
```

---

## Validation Checklist

Before accepting AI-generated code:

- [ ] Runs: `mvn clean test` passes all tests
- [ ] Follows: Matches established patterns from PATTERNS.md
- [ ] Scoped: Only touches declared modules
- [ ] Constrained: Respects performance/architectural constraints
- [ ] Consistent: Uses same naming, structure as surrounding code
- [ ] Documented: New public APIs have JavaDoc
- [ ] Tested: Has accompanying tests (unit or integration)
- [ ] Safe: No SQL injection, no XSS, no unhandled nulls
- [ ] Efficient: No N+1 queries, reasonable complexity

---

## Common Failure Modes & Fixes

| Failure | Root Cause | Fix |
|---------|-----------|-----|
| AI changes unrelated files | No scope definition | Add SCOPE section to prompt |
| Code compiles but fails at runtime | Spring wiring not validated | Run `mvn test`, not just compile |
| AI suggests new patterns | No PATTERNS.md | Document patterns first |
| AI hallucinates method names | No exploration phase | Force `TASK 1: Find X` step |
| Code violates module rules | No MAVEN-FACTS.md | Provide import rules upfront |
| Performance regression | No constraints documented | Add SLA requirements |
| Test coverage drops | No TDD approach | Always write tests first |
| Circular dependencies introduced | No architecture clarity | Use pom-map.txt |

---

## Quantified Impact

### Before (Typical Failure)
- Request → Hallucinated code → Compilation errors → Back-to-back prompts
- **Time: 2-3 hours** for medium task
- **Quality: 60%** (works after fixes)
- **Trust: Low** (need extensive review)

### After (Steroids Applied)
- Prep (5 min) → Explore (5 min) → Implement (10 min) → Validate (10 min) → Commit (5 min)
- **Time: 35 minutes** for same task
- **Quality: 95%** (runs first try)
- **Trust: High** (follows patterns, passes tests)

### ROI Example
- Large refactoring (200 file changes):
  - Without steroids: 40-60 hours
  - With steroids: 8-10 hours
  - **Saving: 30-50 hours per project**

---

## Applicable to All Tools

These strategies work regardless of tool:
- ✅ Gemini 3.1 Pro
- ✅ Claude (any version)
- ✅ Cursor CLI
- ✅ Codex / GPT-4
- ✅ LLaMA-based tools
- ✅ Open-source alternatives

The technique transcends the LLM. Better input → better output, always.

---

## Next Steps

1. **Audit your codebase**:
   - Run: `mvn dependency:tree > pom-map.txt`
   - List all @Service classes
   - List all @Aspect classes
   - Identify 3-5 core patterns

2. **Create foundation docs**:
   - ARCHITECTURE.md (30 min)
   - PATTERNS.md (30 min)
   - MAVEN-FACTS.md (15 min)
   - SPRING-FACTS.md (15 min)

3. **Try with one small task**:
   - Pick: Add logging / Add validation / Refactor small service
   - Follow: Prep → Explore → Implement → Validate → Commit workflow
   - Measure: Time, quality, confidence

4. **Iterate**: Each task, refine your documentation based on AI's questions

---

## Conclusion

Gemini 3.1 Pro (and competitors) aren't subpar—**they're unapplied**. The gap between mediocre and excellent AI-assisted development is not in the tool; it's in how you guide it.

**The steroids work universally because they solve a fundamental problem: reducing ambiguity.**

When an LLM has:
- Clear scope
- Established patterns to follow
- Actual codebase facts (not guesses)
- Incremental milestones with validation
- Test-driven validation

…it produces professional-grade code, first try.

The 10 techniques aren't optional extras. They're the **operating system for effective AI-assisted development** at scale.
