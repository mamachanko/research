# Architecture Template for Large Java/Spring Boot Codebases

Use this template to document your codebase structure for AI tools.

## 1. Module Hierarchy

```
project-root/
├── pom.xml (root, packaging: pom)
│
├── api-contracts/
│   └── pom.xml (packaging: jar, no Spring dependencies)
│       └── src/main/java/com/company/api/
│           ├── UserRequest.java
│           ├── UserResponse.java
│           ├── PaymentRequest.java
│           └── PaymentResponse.java
│
├── common-lib/
│   └── pom.xml (packaging: jar, shared utilities, no Spring)
│       └── src/main/java/com/company/common/
│           ├── exception/
│           │   ├── DomainException.java
│           │   ├── ValidationException.java
│           │   └── PaymentException.java
│           ├── util/
│           │   ├── DateUtil.java
│           │   └── ValidationUtil.java
│           └── model/
│               └── Money.java
│
├── payment-service/
│   └── pom.xml (packaging: jar, Spring Boot app)
│       └── src/main/java/com/company/payment/
│           ├── config/
│           │   └── PaymentConfig.java
│           ├── controller/
│           │   └── PaymentController.java
│           ├── service/
│           │   ├── PaymentService.java
│           │   └── PaymentProcessor.java
│           ├── repository/
│           │   ├── PaymentRepository.java
│           │   └── PaymentStatusRepository.java
│           ├── entity/
│           │   ├── Payment.java
│           │   └── PaymentStatus.java
│           └── dto/
│               ├── PaymentRequest.java
│               └── PaymentResponse.java
│
└── user-service/
    └── pom.xml (packaging: jar, Spring Boot app)
        └── src/main/java/com/company/user/
            ├── config/
            ├── controller/
            ├── service/
            ├── repository/
            ├── entity/
            └── dto/
```

## 2. Module Dependencies

### Import Rules (Enforce These)
```
payment-service CAN import:
  ✓ com.company.common.*
  ✓ com.company.api.*
  ✓ javax.*, java.*
  ✗ com.company.user.*
  ✗ Spring components from other services

user-service CAN import:
  ✓ com.company.common.*
  ✓ com.company.api.*
  ✓ javax.*, java.*
  ✗ com.company.payment.*

common-lib CAN import:
  ✓ Only java.*, javax.*, org.slf4j
  ✗ Spring dependencies
  ✗ Service-specific classes

api-contracts CAN import:
  ✓ Only java.*, javax.*
  ✗ Spring dependencies
  ✗ Any service logic
```

## 3. Layer Structure (Per Service)

```
Web Layer (Controllers)
    ↓ receives HTTP
├── @RestController endpoints
├── Parse request → Request DTOs
├── Call service layer
└── Map response → Response DTOs

    ↓ delegates to
Service Layer (Services & Processors)
├── @Service classes
├── Business logic
├── Transaction boundaries (@Transactional)
├── Validation
├── Call repository layer
└── May call external services

    ↓ delegates to
Persistence Layer (Repositories)
├── Spring Data JpaRepository
├── Custom query methods
├── Fetch/save entities
└── Exception handling

    ↓ manages
Entity Layer (Domain Models)
├── @Entity classes
├── JPA mappings
├── Basic validation
└── Minimal logic

Cross-Cutting (Aspects)
├── @Aspect: Logging
├── @Aspect: Security
├── @Aspect: Audit
└── @Aspect: Metrics
```

## 4. Configuration Sources

### Properties Files (Applied in Order)
```
1. application.properties (base, always loaded)
2. application-{profile}.properties (profile-specific)
   - application-dev.properties
   - application-staging.properties
   - application-prod.properties

Active profile determined by: -Dspring.profiles.active=dev

Example:
# application.properties
spring.datasource.url=jdbc:mysql://localhost:3306/appdb

# application-prod.properties
spring.datasource.url=jdbc:mysql://prod-db:3306/appdb
logging.level.root=WARN
```

### @Configuration Classes (Beans Programmatically)
```
PaymentConfig.java
  ├── @Bean RestTemplate
  ├── @Bean WebClient
  ├── @Bean ObjectMapper
  └── @ConditionalOnProperty customPaymentGateway

AuditConfig.java
  ├── @Bean AuditAspect
  └── @Bean AuditRepository

SecurityConfig.java
  ├── @Bean SecurityFilterChain
  └── @Bean PasswordEncoder
```

## 5. Bean Creation & Wiring

### Auto-Created (by Spring Boot)
```
✓ DataSource (from spring.datasource.*)
✓ JpaTransactionManager
✓ EntityManagerFactory
✓ ServletDispatcher
✓ RestTemplate (if on classpath)
✓ All @Service/@Repository/@Controller classes
✓ All @Bean from @Configuration classes
```

### Explicitly Not Auto-Created
```
✗ Custom PaymentValidator (must @Component or @Service)
✗ External API clients (must @Bean)
✗ Custom thread pools (must @Bean)
```

## 6. Transaction Boundaries

```
Transaction Scope Rules:
├── @Transactional on: Service write methods
│   ├── processPayment()
│   ├── updateStatus()
│   └── refund()
│
├── @Transactional(readOnly=true) on: Query methods
│   ├── findByUserId()
│   ├── getPaymentHistory()
│   └── search()
│
└── No @Transactional on:
    ├── Controllers
    ├── Aspect methods
    └── Async methods (unless needed)
```

## 7. Naming Conventions

| Component | Pattern | Example |
|-----------|---------|---------|
| Entity | [Domain] | Payment, User, Order |
| Service | [Domain]Service | PaymentService |
| Repository | [Domain]Repository | PaymentRepository |
| Controller | [Domain]Controller | PaymentController |
| Request DTO | [Action]Request | PaymentRequest, RefundRequest |
| Response DTO | [Domain]Response | PaymentResponse |
| Config | [Domain]Config | PaymentConfig |
| Exception | [Domain]Exception | PaymentException |
| Aspect | [Concern]Aspect | AuditAspect, LoggingAspect |
| Validator | [Domain]Validator | PaymentValidator |

## 8. Error Handling Strategy

```
Layer: All layers
├── Catch checked exceptions
├── Convert to domain exceptions
│   ├── ValidationException → 400 Bad Request
│   ├── PaymentException → 409 Conflict
│   └── RuntimeException → 500 Internal Server Error
└── Log with context (user ID, payment ID, etc.)

GlobalExceptionHandler
├── @ExceptionHandler(ValidationException)
├── @ExceptionHandler(PaymentException)
└── Returns ApiError DTO
    {
      "code": "INVALID_AMOUNT",
      "message": "Amount must be > 0",
      "timestamp": "2024-01-15T10:30:00Z"
    }
```

## 9. Testing Strategy

| Type | Scope | Setup | Location |
|------|-------|-------|----------|
| Unit | Single class, mocked dependencies | @ExtendWith(MockitoExtension.class) | PaymentServiceTest |
| Integration | Service + Repository, real DB | @SpringBootTest, Testcontainers | PaymentServiceIT |
| Controller | Controller + Service | @WebMvcTest with @MockBean | PaymentControllerTest |
| Acceptance | Full flow, real app context | @SpringBootTest, all real beans | PaymentAcceptanceTest |

## 10. Documentation References

For AI tools, provide:
- [ ] This ARCHITECTURE.md
- [ ] PATTERNS.md (established patterns)
- [ ] MAVEN-FACTS.md (build details)
- [ ] SPRING-FACTS.md (Spring configuration)
- [ ] pom-map.txt (output of `mvn dependency:tree`)

---

## Quick Reference for AI

**When you need to understand the codebase, reference:**
1. Module structure above (section 1)
2. Layer structure (section 3)
3. Import rules (section 2)
4. Naming conventions (section 7)

**When you're implementing something, check:**
1. Is there an existing pattern? (see PATTERNS.md)
2. What layer does this belong in? (see section 3)
3. Does my implementation follow naming conventions? (see section 7)
4. Will this respect import rules? (see section 2)
