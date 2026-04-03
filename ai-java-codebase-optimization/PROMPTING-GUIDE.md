# AI Prompting Guide for Large Java Codebases

Use these prompting templates to maximize quality output from any LLM tool.

---

## Template 1: Exploration Phase

**Goal**: Make AI discover actual patterns before implementing.

```
EXPLORATION TASK
================

Do NOT implement anything yet. Explore the codebase and report findings.

Using the provided ARCHITECTURE.md and PATTERNS.md:

1. SERVICES
   Find all @Service classes. List 5 of them with their responsibilities.
   Example format: "UserService → handles user registration, profile updates"

2. REPOSITORIES  
   List all @Repository interfaces and their methods.
   Example: "PaymentRepository → findById(), findByUserId(), findByStatus()"

3. CUSTOM ANNOTATIONS
   Find all @Target(ElementType.METHOD) annotations (like @Auditable, @Loggable).
   Show one usage example for each.

4. ASPECTS
   Find all @Aspect classes. Show what they intercept and how.

5. EXCEPTION HIERARCHY
   Show all custom exceptions and their parent classes.

6. KEY PATTERNS
   Identify 3 patterns you found repeatedly:
   - Service class pattern
   - Repository pattern
   - DTO pattern

REFERENCE the actual file paths and code snippets in your response.
DO NOT guess or assume—find and reference real code.
```

---

## Template 2: Scope Definition (Before Implementation)

**Goal**: Lock boundaries so AI doesn't scope-creep.

```
IMPLEMENTATION SCOPE
====================

TASK: [What you want done]

SCOPE (only touch these):
  ✓ Modules: payment-service/src/main/java/com/company/payment/
  ✓ Can modify: Service, Repository, DTO classes in payment-service
  ✗ Cannot modify: shared-lib, common-lib, user-service

CONSTRAINTS (do NOT suggest alternatives):
  ✗ Do NOT change existing method signatures
  ✗ Do NOT add new dependencies
  ✗ Do NOT modify transaction boundaries
  ✗ Do NOT change error handling approach
  ✓ Must maintain existing patterns from PATTERNS.md

PERFORMANCE CONSTRAINTS:
  - Response time must remain < 200ms
  - No N+1 queries (max 3 DB queries per request)
  - Memory footprint must not increase

VALIDATION:
  - Code must pass: mvn clean test
  - Must match patterns in PATTERNS.md
  - No new public APIs without JavaDoc
```

---

## Template 3: Pattern-Driven Implementation

**Goal**: AI extends existing patterns instead of creating new ones.

```
PATTERN-DRIVEN IMPLEMENTATION
=============================

Review PATTERNS.md before proceeding.

TASK: [What you want done]

Required Pattern References:
1. Follow Service Layer Pattern (constructor injection, @Transactional)
2. Follow Repository Pattern (one repo per entity, @Query for complex)
3. Follow Error Handling Pattern (custom exceptions, no catch-and-return)
4. Follow Validation Pattern (separate Validator class)
5. Follow Testing Pattern (unit tests mock, integration tests Testcontainers)

IMPLEMENTATION STEPS:
1. Write test shape first (signatures only)
2. Implement the feature
3. Ensure tests pass
4. Verify patterns followed

CHECKLIST before returning:
  [ ] Constructor injection used (no @Autowired)
  [ ] @Transactional on write operations only
  [ ] DTOs returned, not entities
  [ ] Custom exceptions thrown, not raw exceptions
  [ ] Tests written and passing
  [ ] No code duplication with existing patterns
```

---

## Template 4: Incremental Multi-Step Tasks

**Goal**: Break complex tasks into milestones with checkpoints.

```
MULTI-STEP IMPLEMENTATION
=========================

FINAL GOAL: [Complex refactoring/feature]

This is a 5-step task. Complete ONE step, then wait for approval before next.

STEP 1: [Specific sub-task]
  - What to explore/find/understand
  - Deliverable: [specific output]
  - No implementation yet

STEP 2: [Next sub-task]
  - Depends on Step 1
  - Deliverable: [specific output]

STEP 3: [Implementation sub-task]
  - Now implement based on Steps 1-2
  - Must pass tests
  - Deliverable: [specific output]

STEP 4: [Integration sub-task]
  - Connect with existing code
  - Deliverable: [specific output]

STEP 5: [Validation sub-task]
  - Run full test suite
  - Performance check
  - Deliverable: [specific output]

CHECKPOINT: After each step, I will review and approve before proceeding.
```

---

## Template 5: Few-Shot Learning (Teach AI Your Pattern)

**Goal**: Show by example, then apply to new code.

```
EXAMPLE-DRIVEN IMPLEMENTATION
==============================

TASK: [What you want done]

First, here's how we implement Services in this codebase:

[PASTE REAL SERVICE EXAMPLE - UserService or similar]

And here's how we test Services:

[PASTE REAL TEST EXAMPLE - UserServiceTest]

Now implement [New Service] following this exact pattern:
  - Same injection style
  - Same method structure
  - Same error handling
  - Same transaction boundaries
  - Same test structure
  
Use only the patterns shown in examples above.
Do NOT suggest alternatives or improvements to the pattern.
```

---

## Template 6: Architecture Verification

**Goal**: Ensure AI respects module boundaries and import rules.

```
ARCHITECTURE-AWARE IMPLEMENTATION
==================================

TASK: [What you want done]

Reference pom-map.txt for module dependencies.

IMPORT RULES (strictly enforced):
  payment-service imports from: common-lib, api-contracts
  user-service imports from: common-lib, api-contracts
  common-lib imports from: nothing else

VERIFICATION:
  1. What module are you modifying? [payment-service]
  2. What modules can you import? [common-lib, api-contracts]
  3. Will any imports violate rules? [Check before implementing]

Before submitting code, verify:
  [ ] All imports come from allowed modules
  [ ] No circular dependencies introduced
  [ ] No cross-service coupling (payment ↔ user)
```

---

## Template 7: Test-First Development

**Goal**: Let tests drive implementation.

```
TEST-FIRST IMPLEMENTATION
=========================

TASK: [Implement feature X]

STEP 1: Write test signatures (no implementation)
  Write these test methods (SIGNATURES ONLY):
  - testFeatureX_WithValidInput_Succeeds()
  - testFeatureX_WithInvalidInput_ThrowsException()
  - testFeatureX_WithNullDependency_HandlesGracefully()
  
  Use structure from PATTERNS.md test examples.

STEP 2: Write tests (full implementations)
  Implement the test methods.
  Do NOT implement feature yet.
  Run tests; they should fail (red).

STEP 3: Implement feature
  Implement feature to make tests pass (green).
  
STEP 4: Verify
  Ensure all tests pass: mvn clean test

ALWAYS test-first. Tests clarify intent better than requirements.
```

---

## Template 8: Domain-Driven Implementation

**Goal**: AI generates semantically correct domain models.

```
DOMAIN-DRIVEN IMPLEMENTATION
============================

TASK: [Implement domain feature]

Domain Model (DDD):
  Aggregate: [Name]
    - Identity: [field]
    - State: [fields + invariants]
    - Commands: [valid transitions]
    - Events: [domain events emitted]

  Example for Payment aggregate:
    - Identity: paymentId
    - State: status (PENDING → PROCESSING → COMPLETED)
    - Invariants: amount > 0, status transitions valid
    - Commands: processPayment(), refund(), updateStatus()
    - Events: PaymentProcessedEvent, PaymentFailedEvent

Value Objects:
  - Money (amount + currency)
  - PaymentStatus (with valid transitions)

Repositories:
  - PaymentRepository (persistence)

Services:
  - PaymentProcessor (orchestration, external integrations)

IMPLEMENTATION:
  1. Create domain model respecting aggregates/invariants
  2. Create repository for persistence
  3. Create service for orchestration
  4. Emit domain events on state changes

Use DDD language: aggregates, repositories, domain events, value objects.
DO NOT use: anemic models, data classes, no invariant validation.
```

---

## Template 9: Performance-Aware Implementation

**Goal**: AI considers performance constraints.

```
PERFORMANCE-CONSTRAINED IMPLEMENTATION
=======================================

TASK: [What you want done]

PERFORMANCE REQUIREMENTS:
  - HTTP response time: < 200ms (p99)
  - Database queries per request: max 3
  - Memory footprint: must not increase > 10%
  - CPU usage: must not increase > 5%

CONSTRAINTS:
  - No N+1 queries (use @Query with JOIN FETCH)
  - No unnecessary database calls
  - Cache frequently accessed data if beneficial
  - Batch bulk operations

VALIDATION:
  Before submitting, verify:
    [ ] Query count: find all DB calls in method
    [ ] Join patterns: check for N+1
    [ ] Caching: identify caching opportunities
    [ ] Batch operations: look for bulk loops

Code review will include performance check.
Queries that fail will require rework.
```

---

## Template 10: Documentation Verification

**Goal**: New code includes proper documentation.

```
DOCUMENTATION-REQUIRED IMPLEMENTATION
======================================

TASK: [What you want done]

All new public methods/classes must have:

1. JavaDoc on public classes:
   /**
    * Brief description.
    * 
    * Longer description if needed.
    */

2. JavaDoc on public methods:
   /**
    * Brief description.
    * 
    * @param paramName description
    * @return description
    * @throws ExceptionType when condition
    */

3. Complex logic comments:
   // Explain WHY, not WHAT
   // Example: "Prevent N+1 by using JOIN FETCH"

4. No documentation on:
   - Getters/setters (obvious)
   - Overridden methods (inherit from parent)
   - Private methods (use clear names instead)

Submission must include proper documentation.
Missing JavaDoc will be rejected.
```

---

## Workflow: How to Use These Templates

### For Small Tasks (< 2 hours)
1. Use **Template 2** (Scope Definition)
2. Use **Template 3** (Pattern-Driven)
3. Use **Template 7** (Test-First)
4. Submit

### For Medium Tasks (2-8 hours)
1. Use **Template 1** (Exploration)
2. Use **Template 5** (Few-Shot)
3. Use **Template 4** (Multi-Step)
4. Use **Template 7** (Test-First)
5. Use **Template 9** (Performance)
6. Submit

### For Large Refactorings (8+ hours)
1. Use **Template 1** (Exploration)
2. Use **Template 6** (Architecture)
3. Use **Template 4** (Multi-Step with many steps)
4. Use **Template 8** (Domain-Driven)
5. Use **Template 7** (Test-First for each step)
6. Use **Template 9** (Performance)
7. Use **Template 10** (Documentation)
8. Submit in chunks with reviews

---

## Anti-Patterns in Prompting

### ❌ What NOT to do:

1. **No Scope**: "Refactor the payment module"
   - **Better**: "Add audit logging to PaymentService.processPayment(). Do NOT touch entity, repository, or DTOs."

2. **No Context**: "Implement X feature"
   - **Better**: "Based on ARCHITECTURE.md and PATTERNS.md, implement X feature following [pattern name]"

3. **No Constraints**: "Optimize this code"
   - **Better**: "Optimize response time to < 100ms, max 3 queries, no new dependencies"

4. **No Exploration**: "Add caching here"
   - **Better**: "TASK 1: Find all slow queries. TASK 2: Design cache strategy. TASK 3: Implement caching."

5. **Mixed Concerns**: "Refactor service, add tests, deploy"
   - **Better**: "Refactor service. [Wait for approval]. Then add tests. [Wait]. Then deploy."

6. **No Validation**: "Here's the code, use it"
   - **Better**: "Here's the code. Run: mvn test. If tests fail, fix. Then use it."

7. **Assume Patterns**: "Make it DDD"
   - **Better**: "Show me 3 existing domain models. Then implement X as domain aggregate with invariants."

---

## Prompting Checklists

### Before Every Request, Ask:

```
✓ Is the scope clear? (what to modify, what not to)
✓ Are constraints explicit? (no refactoring, no new deps, etc.)
✓ Are patterns documented? (PATTERNS.md provided)
✓ Is architecture clear? (ARCHITECTURE.md provided)
✓ Is exploration a first step? (or obvious enough to skip)
✓ Are tests included? (test-first approach)
✓ Is validation criteria defined? (how to know it's done)
✓ Is incremental approach used? (for complex tasks)
✓ Are checkpoints defined? (where to pause and review)
```

If all checked, submit prompt. If any missing, add before asking.

