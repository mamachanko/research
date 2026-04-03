# Established Patterns Template

Copy this template and fill in your actual patterns. Share with AI before requesting implementations.

## Service Layer Pattern

### Rule: Always Constructor Injection
```java
@Service
public class PaymentService {
    private final PaymentRepository repository;
    private final PaymentValidator validator;
    private final AuditService auditService;

    // Constructor, no field injection
    public PaymentService(PaymentRepository repository, 
                          PaymentValidator validator,
                          AuditService auditService) {
        this.repository = repository;
        this.validator = validator;
        this.auditService = auditService;
    }
}

// DO NOT do:
@Service
public class BadPaymentService {
    @Autowired PaymentRepository repository;  // ❌ NO
}
```

### Rule: @Transactional on Write Operations
```java
@Service
public class PaymentService {
    
    @Transactional
    public PaymentResponse processPayment(PaymentRequest req) {
        // Write operation: must have @Transactional
    }

    @Transactional
    public void updatePaymentStatus(Long paymentId, PaymentStatus status) {
        // Write operation
    }

    @Transactional(readOnly=true)
    public PaymentResponse getPayment(Long paymentId) {
        // Read operation: use readOnly=true for optimization
        return repository.findById(paymentId)
            .map(this::toResponse)
            .orElseThrow(...);
    }

    @Transactional(readOnly=true)
    public List<PaymentResponse> getPaymentHistory(Long userId) {
        // Query method: readOnly=true
        return repository.findByUserId(userId).stream()
            .map(this::toResponse)
            .toList();
    }
}
```

### Rule: No Entity Returns; Always Use DTOs
```java
@Service
public class PaymentService {
    
    @Transactional(readOnly=true)
    public PaymentResponse getPayment(Long paymentId) {
        Payment entity = repository.findById(paymentId)
            .orElseThrow(() -> new PaymentNotFoundException(...));
        
        return toResponse(entity);  // Convert to DTO before returning
    }

    // Conversion helper (private)
    private PaymentResponse toResponse(Payment entity) {
        return new PaymentResponse(
            entity.getId(),
            entity.getAmount(),
            entity.getStatus(),
            entity.getCreatedAt()
        );
    }

    // DO NOT return entity directly:
    // ❌ public Payment getPayment(Long id) { ... }
}
```

---

## Repository Pattern

### Rule: One Repository per Entity
```java
@Repository
public interface PaymentRepository extends JpaRepository<Payment, Long> {
    List<Payment> findByUserId(Long userId);
    List<Payment> findByStatus(PaymentStatus status);
    Optional<Payment> findByIdempotencyKey(String key);
}

// One repository, multiple query methods
// NOT multiple repositories for same entity
```

### Rule: Custom Queries with @Query
```java
@Repository
public interface PaymentRepository extends JpaRepository<Payment, Long> {
    
    @Query("SELECT p FROM Payment p WHERE p.userId = :userId AND p.status = :status ORDER BY p.createdAt DESC")
    List<Payment> findByUserAndStatus(@Param("userId") Long userId, 
                                      @Param("status") PaymentStatus status);

    @Query("SELECT COUNT(p) FROM Payment p WHERE p.userId = :userId")
    int countByUserId(@Param("userId") Long userId);
}

// Use @Query for complex queries, simple methods for simple queries
```

---

## Error Handling Pattern

### Rule: Custom Exceptions
```java
// Domain exception hierarchy
public class DomainException extends RuntimeException {
    public DomainException(String message) {
        super(message);
    }
}

public class PaymentException extends DomainException {
    public PaymentException(String message) {
        super(message);
    }
}

public class ValidationException extends DomainException {
    public ValidationException(String message) {
        super(message);
    }
}

public class PaymentNotFoundException extends PaymentException {
    public PaymentNotFoundException(Long id) {
        super("Payment not found: " + id);
    }
}
```

### Rule: Throw in Service, Handle in Aspect/Controller
```java
@Service
public class PaymentService {
    
    @Transactional
    public PaymentResponse processPayment(PaymentRequest req) {
        // Validate early
        if (req.getAmount() <= 0) {
            throw new ValidationException("Amount must be > 0");
        }

        Payment payment = repository.findById(req.getPaymentId())
            .orElseThrow(() -> new PaymentNotFoundException(req.getPaymentId()));

        // Process
        return toResponse(payment);
    }
}

@RestController
@RequestMapping("/api/payments")
public class PaymentController {
    
    @PostMapping
    public ResponseEntity<ApiResponse> process(@RequestBody PaymentRequest req) {
        try {
            PaymentResponse resp = service.processPayment(req);
            return ResponseEntity.ok(ApiResponse.success(resp));
        } catch (ValidationException e) {
            return ResponseEntity
                .badRequest()
                .body(ApiResponse.error("VALIDATION_ERROR", e.getMessage()));
        } catch (PaymentException e) {
            return ResponseEntity
                .status(409)
                .body(ApiResponse.error("PAYMENT_ERROR", e.getMessage()));
        }
    }
}

// DO NOT catch exceptions in service and return error objects
// ❌ public PaymentResponse processPayment(req) {
//      try { ... } catch { return error(); }
//    }
```

---

## Validation Pattern

### Rule: Separate Validator Class
```java
@Service
public class PaymentValidator {
    
    public void validate(PaymentRequest request) throws ValidationException {
        validateAmount(request.getAmount());
        validateCurrency(request.getCurrency());
        validateUser(request.getUserId());
    }

    private void validateAmount(BigDecimal amount) {
        if (amount == null || amount.compareTo(BigDecimal.ZERO) <= 0) {
            throw new ValidationException("Amount must be > 0");
        }
        if (amount.compareTo(new BigDecimal("1000000")) > 0) {
            throw new ValidationException("Amount exceeds limit");
        }
    }

    private void validateCurrency(String currency) {
        if (!Set.of("USD", "EUR", "GBP").contains(currency)) {
            throw new ValidationException("Unsupported currency: " + currency);
        }
    }

    private void validateUser(Long userId) {
        if (userId == null || userId <= 0) {
            throw new ValidationException("Invalid user ID");
        }
    }
}

@Service
public class PaymentService {
    private final PaymentValidator validator;

    @Transactional
    public PaymentResponse processPayment(PaymentRequest req) {
        validator.validate(req);  // Validate early
        // ... rest of logic
    }
}
```

---

## Aspect-Based Cross-Cutting Concerns

### Rule: Use @Aspect for Logging, Audit, Metrics
```java
@Aspect
@Component
public class LoggingAspect {
    
    private static final Logger log = LoggerFactory.getLogger(LoggingAspect.class);

    @Around("@annotation(Loggable)")
    public Object log(ProceedingJoinPoint jp) throws Throwable {
        String method = jp.getSignature().getName();
        long start = System.currentTimeMillis();

        try {
            Object result = jp.proceed();
            long duration = System.currentTimeMillis() - start;
            log.info("Method {} completed in {}ms", method, duration);
            return result;
        } catch (Exception e) {
            long duration = System.currentTimeMillis() - start;
            log.error("Method {} failed after {}ms", method, duration, e);
            throw e;
        }
    }
}

// Usage
@Service
public class PaymentService {
    
    @Loggable
    @Transactional
    public PaymentResponse processPayment(PaymentRequest req) {
        // Logic automatically logged
    }
}

// DO NOT use decorators or manual logging:
// ❌ public PaymentResponse processPayment(...) {
//      log.info("Starting process");
//      try { ... } finally { log.info("Done"); }
//    }
```

---

## Testing Pattern

### Rule: Unit Tests Mock All Dependencies
```java
@ExtendWith(MockitoExtension.class)
class PaymentServiceTest {
    
    @Mock PaymentRepository repository;
    @Mock PaymentValidator validator;
    @InjectMocks PaymentService service;

    @Test
    void processPayment_WithValidAmount_Succeeds() {
        // Arrange
        PaymentRequest request = new PaymentRequest(100.0, "USD", 1L);
        Payment entity = new Payment(1L, 100.0, "USD", PENDING);
        
        when(repository.findById(1L)).thenReturn(Optional.of(entity));
        doNothing().when(validator).validate(request);

        // Act
        PaymentResponse response = service.processPayment(request);

        // Assert
        assertThat(response.getId()).isEqualTo(1L);
        assertThat(response.getStatus()).isEqualTo(PENDING);
        verify(repository).findById(1L);
        verify(validator).validate(request);
    }

    @Test
    void processPayment_WithInvalidAmount_ThrowsException() {
        // Arrange
        PaymentRequest request = new PaymentRequest(-100.0, "USD", 1L);
        doThrow(new ValidationException("Amount must be > 0"))
            .when(validator).validate(request);

        // Act & Assert
        assertThatThrownBy(() -> service.processPayment(request))
            .isInstanceOf(ValidationException.class)
            .hasMessage("Amount must be > 0");
    }
}
```

### Rule: Integration Tests Use Testcontainers
```java
@SpringBootTest
@Testcontainers
class PaymentServiceIT {
    
    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>()
        .withDatabaseName("testdb")
        .withUsername("test")
        .withPassword("test");

    @Autowired PaymentService service;
    @Autowired PaymentRepository repository;

    @Test
    void processPayment_WithRealDatabase_Persists() {
        // Arrange
        PaymentRequest request = new PaymentRequest(100.0, "USD", 1L);

        // Act
        PaymentResponse response = service.processPayment(request);

        // Assert
        Payment saved = repository.findById(response.getId()).orElseThrow();
        assertThat(saved.getAmount()).isEqualTo(100.0);
        assertThat(saved.getStatus()).isEqualTo(PENDING);
    }
}

// DO NOT use H2 or in-memory DB
// ❌ Use spring.datasource.url=jdbc:h2:mem:testdb
```

---

## DTO Pattern

### Rule: Request/Response DTOs Immutable
```java
@Data
@AllArgsConstructor
@NoArgsConstructor
public class PaymentRequest {
    private Long paymentId;
    private BigDecimal amount;
    private String currency;
    private Long userId;
    
    // Validation in constructor or separate validator
}

@Data
@Builder
public class PaymentResponse {
    private Long id;
    private BigDecimal amount;
    private String currency;
    private PaymentStatus status;
    private Instant createdAt;
    
    // NO setters, immutable
}

// DO NOT use mutable getters/setters in DTOs:
// ❌ public void setAmount(BigDecimal amount) { }
```

---

## Configuration Pattern

### Rule: External Configuration in @Configuration Classes
```java
@Configuration
public class PaymentConfig {
    
    @Bean
    public RestTemplate restTemplate() {
        return new RestTemplate();
    }

    @Bean
    @ConditionalOnProperty(name = "payment.async.enabled", havingValue = "true")
    public PaymentAsyncProcessor asyncPaymentProcessor() {
        return new PaymentAsyncProcessor();
    }
}

// Use properties to inject configuration
@Service
public class PaymentService {
    @Value("${payment.max-retries:3}")
    private int maxRetries;

    @Value("${payment.timeout-ms:5000}")
    private long timeoutMs;
}
```

---

## Summary: Pattern Checklist

When implementing, verify:
- [ ] Constructor injection (not @Autowired)
- [ ] @Transactional on write, @Transactional(readOnly=true) on read
- [ ] DTOs for all returns, never entities
- [ ] Custom exceptions, no raw exceptions
- [ ] Validation in separate Validator class
- [ ] Aspects for cross-cutting concerns (logging, audit)
- [ ] Unit tests mock all dependencies
- [ ] Integration tests use Testcontainers
- [ ] Immutable DTOs (@Data, no setters)
- [ ] @Configuration for external beans

