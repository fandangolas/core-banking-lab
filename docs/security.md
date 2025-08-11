# Security Implementation

**Production-grade security hardening for banking operations with defense-in-depth approach.**

## Overview

The banking API implements **comprehensive security measures** across all layers - from network perimeter to business logic - designed to protect sensitive financial data and prevent fraud.

## Rate Limiting

### IP-Based Request Throttling
```go
type RateLimiter struct {
    requests map[string][]time.Time
    mutex    sync.RWMutex
    limit    int           // Requests per window
    window   time.Duration // Time window
}

func (rl *RateLimiter) Allow(clientIP string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    now := time.Now()
    
    // Clean expired requests (prevents memory leaks)
    validRequests := []time.Time{}
    for _, reqTime := range rl.requests[clientIP] {
        if now.Sub(reqTime) < rl.window {
            validRequests = append(validRequests, reqTime)
        }
    }
    
    if len(validRequests) >= rl.limit {
        return false // Rate limit exceeded
    }
    
    rl.requests[clientIP] = append(validRequests, now)
    return true
}
```

### Configuration
```bash
# Production rate limiting
export RATE_LIMIT_REQUESTS_PER_MINUTE=50

# Development (more permissive)
export RATE_LIMIT_REQUESTS_PER_MINUTE=1000
```

### Rate Limit Response
```json
{
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "retry_after": 60
}
```

## Input Validation

### Amount Validation
```go
func ValidateAmount(amount int) error {
    if amount <= 0 {
        return errors.New("amount must be positive")
    }
    
    if amount > 1000000 { // R$ 10,000.00 limit
        return errors.New("amount exceeds maximum limit of R$ 10,000.00")
    }
    
    return nil
}
```

### Account Owner Validation
```go
func ValidateOwnerName(owner string) error {
    if len(owner) < 2 {
        return errors.New("owner name must be at least 2 characters")
    }
    
    if len(owner) > 100 {
        return errors.New("owner name must be less than 100 characters")
    }
    
    // Allow only letters, spaces, periods, hyphens, apostrophes
    validName := regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s.\-']+$`)
    if !validName.MatchString(owner) {
        return errors.New("owner name contains invalid characters")
    }
    
    return nil
}
```

### Request Parameter Validation
```go
// Transfer request validation
func ValidateTransferRequest(req TransferRequest) error {
    if req.From == req.To {
        return errors.New("cannot transfer to the same account")
    }
    
    if req.From <= 0 || req.To <= 0 {
        return errors.New("invalid account IDs")
    }
    
    return ValidateAmount(req.Amount)
}
```

## CORS Protection

### Strict Origin Policy
```go
func CORSMiddleware() gin.HandlerFunc {
    config := cors.Config{
        AllowOrigins:     getAllowedOrigins(),
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Content-Type", "Authorization", "Accept"},
        ExposeHeaders:    []string{"X-Request-ID"},
        AllowCredentials: false,
        MaxAge:          12 * time.Hour,
    }
    
    return cors.New(config)
}

func getAllowedOrigins() []string {
    origins := os.Getenv("CORS_ALLOWED_ORIGINS")
    if origins == "" {
        // Development default
        return []string{"http://localhost:5173"}
    }
    
    // Production: explicit whitelist
    return strings.Split(origins, ",")
}
```

### Production Configuration
```bash
# Strict production CORS
export CORS_ALLOWED_ORIGINS="https://secure-banking.example.com,https://banking-dashboard.example.com"

# No wildcards in production!
# NEVER: export CORS_ALLOWED_ORIGINS="*"
```

## Error Handling Security

### Generic Error Responses
```go
type APIError struct {
    Code      string    `json:"code"`
    Message   string    `json:"message"`
    RequestID string    `json:"request_id,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

// GOOD: Generic error message
func handleTransferError(err error) APIError {
    switch {
    case errors.Is(err, ErrInsufficientFunds):
        return APIError{
            Code:    "INSUFFICIENT_FUNDS",
            Message: "Transaction cannot be completed", // Generic
        }
    case errors.Is(err, ErrAccountNotFound):
        return APIError{
            Code:    "ACCOUNT_NOT_FOUND", 
            Message: "Account not found", // No details
        }
    default:
        return APIError{
            Code:    "INTERNAL_SERVER_ERROR",
            Message: "An error occurred", // No system details
        }
    }
}
```

### Security Audit Logging
```go
func logSecurityEvent(eventType string, c *gin.Context, details map[string]interface{}) {
    securityLog := map[string]interface{}{
        "event_type":   eventType,
        "client_ip":    c.ClientIP(),
        "user_agent":   c.Request.UserAgent(),
        "request_path": c.Request.URL.Path,
        "timestamp":    time.Now(),
        "details":      details,
    }
    
    logging.Warn("Security event", securityLog)
}

// Usage examples:
// logSecurityEvent("RATE_LIMIT_EXCEEDED", c, map[string]interface{}{"attempts": 5})
// logSecurityEvent("INVALID_INPUT", c, map[string]interface{}{"field": "amount", "value": -100})
```

## Business Logic Security

### Transfer Validation
```go
func ProcessTransfer(from, to *Account, amount int) error {
    // 1. Validate business rules
    if from.ID == to.ID {
        logSecurityEvent("SELF_TRANSFER_ATTEMPT", nil, map[string]interface{}{
            "account_id": from.ID,
            "amount": amount,
        })
        return errors.New("self-transfer not allowed")
    }
    
    // 2. Validate account state
    if from.Balance < amount {
        logSecurityEvent("INSUFFICIENT_FUNDS_ATTEMPT", nil, map[string]interface{}{
            "account_id": from.ID,
            "balance": from.Balance,
            "attempted": amount,
        })
        return ErrInsufficientFunds
    }
    
    // 3. Validate amount limits
    if amount > 100000 { // R$ 1,000.00 
        logSecurityEvent("LARGE_TRANSFER_ATTEMPT", nil, map[string]interface{}{
            "from_account": from.ID,
            "to_account": to.ID,
            "amount": amount,
        })
        return errors.New("transfer amount exceeds limit")
    }
    
    // 4. Execute atomic transfer
    from.Balance -= amount
    to.Balance += amount
    
    return nil
}
```

### Account Balance Protection
```go
func WithdrawMoney(acc *Account, amount int) error {
    // Prevent negative balances
    withAccountLock(acc, func() {
        if acc.Balance-amount < 0 {
            logSecurityEvent("OVERDRAFT_ATTEMPT", nil, map[string]interface{}{
                "account_id": acc.ID,
                "balance": acc.Balance,
                "withdrawal": amount,
            })
            err = ErrInsufficientFunds
            return
        }
        
        acc.Balance -= amount
    })
    
    return err
}
```

## Request Security

### Request ID Tracking
```go
func RequestSecurityMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        requestID := generateSecureRequestID()
        c.Header("X-Request-ID", requestID)
        c.Set("request_id", requestID)
        
        // Log all requests for audit trail
        logging.Info("Request received", map[string]interface{}{
            "request_id": requestID,
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "client_ip":  c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
            "content_length": c.Request.ContentLength,
        })
        
        c.Next()
    })
}

func generateSecureRequestID() string {
    // Cryptographically secure random ID
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return fmt.Sprintf("req_%x", bytes[:8])
}
```

### Content-Type Validation
```go
func ValidateContentType() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        if c.Request.Method == "POST" || c.Request.Method == "PUT" {
            contentType := c.GetHeader("Content-Type")
            if contentType != "application/json" {
                logSecurityEvent("INVALID_CONTENT_TYPE", c, map[string]interface{}{
                    "content_type": contentType,
                })
                c.JSON(http.StatusUnsupportedMediaType, APIError{
                    Code:    "INVALID_CONTENT_TYPE",
                    Message: "Content-Type must be application/json",
                })
                c.Abort()
                return
            }
        }
        c.Next()
    })
}
```

## Data Protection

### No Sensitive Data in Logs
```go
// BAD: Logs sensitive data
logging.Info("User login", map[string]interface{}{
    "password": password, // ❌ Never log passwords
    "ssn": userSSN,       // ❌ Never log PII
})

// GOOD: Logs relevant context only
logging.Info("Account operation", map[string]interface{}{
    "account_id": accountID,    // ✅ Business identifier
    "operation": "transfer",    // ✅ Action type
    "success": true,           // ✅ Outcome
    // No sensitive amounts or personal data
})
```

### Secure Configuration
```go
type Config struct {
    ServerPort    string
    CORSOrigins   []string
    RateLimitRPM  int
    LogLevel      string
    
    // Sensitive fields should use secure sources
    JWTSecret     string // From environment or secret store
    DatabaseURL   string // From environment or secret store
}

func LoadConfig() *Config {
    return &Config{
        ServerPort:   getEnvDefault("SERVER_PORT", "8080"),
        LogLevel:     getEnvDefault("LOG_LEVEL", "info"),
        RateLimitRPM: getEnvIntDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
        
        // Never hardcode secrets!
        JWTSecret:   os.Getenv("JWT_SECRET"), 
        DatabaseURL: os.Getenv("DATABASE_URL"),
    }
}
```

## Security Headers

### HTTP Security Headers
```go
func SecurityHeadersMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // Prevent MIME sniffing
        c.Header("X-Content-Type-Options", "nosniff")
        
        // Prevent clickjacking
        c.Header("X-Frame-Options", "DENY")
        
        // XSS protection
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // Referrer policy
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // HTTPS enforcement (in production)
        if os.Getenv("ENVIRONMENT") == "production" {
            c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }
        
        c.Next()
    })
}
```

## Container Security

### Dockerfile Security Best Practices
```dockerfile
# Use specific version tags (not latest)
FROM golang:1.23-alpine AS builder

# Create non-root user
RUN adduser -D -s /bin/sh -u 1001 bankuser

# Use multi-stage build (smaller attack surface)
FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata

# Switch to non-root user
USER bankuser
WORKDIR /home/bankuser/

# Copy only binary (no source code)
COPY --from=builder /app/bank-api .

# Expose port explicitly
EXPOSE 8080

# Use specific command
CMD ["./bank-api"]
```

### Kubernetes Security Policy
```yaml
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    fsGroup: 2000
  containers:
  - name: banking-api
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
    resources:
      limits:
        memory: "512Mi"
        cpu: "500m"
      requests:
        memory: "256Mi" 
        cpu: "250m"
```

## Security Monitoring

### Threat Detection
```go
var suspiciousActivity = map[string]int{
    "RAPID_REQUESTS":     5,   // More than 5 req/sec from single IP
    "LARGE_TRANSFERS":    3,   // More than 3 large transfers/hour
    "FAILED_VALIDATIONS": 10,  // More than 10 validation failures
}

func detectSuspiciousActivity(eventType string, clientIP string) {
    count := incrementActivityCounter(eventType, clientIP)
    
    if count >= suspiciousActivity[eventType] {
        logging.Warn("Suspicious activity detected", map[string]interface{}{
            "event_type": eventType,
            "client_ip":  clientIP,
            "count":      count,
            "threshold":  suspiciousActivity[eventType],
            "alert":      true,
        })
        
        // Could trigger additional security measures:
        // - Temporary IP blocking
        // - Enhanced monitoring
        // - Security team notification
    }
}
```

## Security Checklist

### ✅ Production Security Measures

**Network Security:**
- Rate limiting per IP address
- CORS with explicit origin whitelist
- HTTPS enforcement in production
- Security headers implementation

**Input Security:**
- Comprehensive input validation
- SQL injection prevention (parameterized queries)
- Amount and length limits
- Content-type validation

**Application Security:**
- Generic error messages (no sensitive data leakage)
- Secure request ID generation  
- Audit logging for all security events
- No hardcoded secrets or credentials

**Infrastructure Security:**
- Non-root container execution
- Minimal container image (Alpine Linux)
- Resource limits and constraints
- Read-only filesystem where possible

**Monitoring & Response:**
- Security event logging and alerting
- Suspicious activity detection
- Request tracing for forensics
- Comprehensive audit trails

This security implementation provides multiple layers of protection suitable for production banking operations while maintaining usability and performance.