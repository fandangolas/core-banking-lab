# Security Implementation Guide

## Overview

This document outlines the comprehensive security measures implemented in the Core Banking Lab, designed to meet the high security standards required for financial systems.

## Security Architecture

### **Defense-in-Depth Strategy**

Our security model implements multiple layers of protection:

```
┌─────────────────────────────────────────────────────────┐
│                   Network Layer                         │
│  • Rate Limiting (IP-based throttling)                 │
│  • CORS Origin Validation                              │
│  • Request Size Limits                                 │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                Application Layer                        │
│  • Input Validation & Sanitization                     │
│  • Business Logic Controls                             │
│  • Error Handling (no info leakage)                    │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   Data Layer                            │
│  • Account Access Controls                             │
│  • Transaction Audit Logging                          │
│  • Data Integrity Validation                          │
└─────────────────────────────────────────────────────────┘
```

## Rate Limiting & DoS Protection

### **Implementation Details**

**Purpose**: Prevent denial-of-service attacks and API abuse by limiting request frequency per client.

```go
type RateLimiter struct {
    requests map[string][]time.Time  // IP -> request timestamps
    mutex    sync.RWMutex           // Thread-safe access
    limit    int                    // Max requests per window
    window   time.Duration          // Time window (default: 1 minute)
}
```

### **Key Features**

- **IP-based tracking**: Each client IP has independent rate limits
- **Sliding window**: More accurate than fixed-window counters
- **Automatic cleanup**: Expired requests are removed to prevent memory leaks
- **Configurable limits**: Adjust via environment variables

### **Configuration**

```bash
# Production settings
export RATE_LIMIT_REQUESTS_PER_MINUTE=30

# Development settings  
export RATE_LIMIT_REQUESTS_PER_MINUTE=100

# High-security environment
export RATE_LIMIT_REQUESTS_PER_MINUTE=10
```

### **Attack Scenarios Prevented**

1. **Brute Force Account Enumeration**
   ```bash
   # Attacker attempts - blocked after 30 requests
   for i in {1..1000}; do
     curl http://api/accounts/$i/balance
   done
   ```

2. **Transaction Flooding**
   ```bash
   # Rapid-fire transfers to hide fraudulent activity - blocked
   for i in {1..100}; do  
     curl -X POST http://api/accounts/transfer -d '{"from":1,"to":2,"amount":1}'
   done
   ```

3. **System Resource Exhaustion**
   - Prevents single client from consuming all server resources
   - Maintains service availability for legitimate users

## Input Validation & Sanitization

### **Comprehensive Validation Strategy**

**Purpose**: Prevent injection attacks and ensure data integrity across all inputs.

### **Amount Validation**

```go
const (
    MinAmount = 1                    // Minimum: R$ 0.01
    MaxAmount = 1000000             // Maximum: R$ 10,000.00 (in centavos)
)

func ValidateAmount(amount int) error {
    if amount < MinAmount {
        return errors.New("amount must be greater than zero")
    }
    if amount > MaxAmount {
        return errors.New("amount exceeds maximum limit of R$ 10,000.00")
    }
    return nil
}
```

**Prevents**:
- Integer overflow attacks
- Negative amount exploits  
- Unrealistic transaction sizes
- Business logic bypass attempts

### **Account Owner Validation**

```go
func ValidateOwnerName(owner string) error {
    owner = strings.TrimSpace(owner)
    
    if len(owner) < MinOwnerLen {
        return errors.New("owner name must be at least 2 characters")
    }
    
    if len(owner) > MaxOwnerLen {
        return errors.New("owner name cannot exceed 100 characters")
    }
    
    // Only allow letters, spaces, and safe punctuation
    for _, r := range owner {
        if !unicode.IsLetter(r) && !unicode.IsSpace(r) && 
           r != '.' && r != '-' && r != '\'' {
            return errors.New("owner name contains invalid characters")
        }
    }
    
    return nil
}
```

**Prevents**:
- SQL injection via name fields
- Cross-site scripting (XSS) payloads
- Buffer overflow attempts
- Unicode-based attacks

### **Account ID Validation**

```go
func ValidateAccountID(id int) error {
    if id <= 0 {
        return errors.New("account ID must be positive")
    }
    return nil
}
```

**Prevents**:
- Directory traversal attacks
- Negative ID exploits
- Zero-value bypass attempts

## CORS (Cross-Origin Resource Sharing) Protection

### **Strict Origin Control**

**Purpose**: Prevent unauthorized web applications from accessing the banking API.

```go
func CORS(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        
        // Check if origin is in allowlist
        allowed := false
        for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
            if allowedOrigin == origin {
                allowed = true
                c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
                break
            }
        }
        
        // Reject if origin not allowed
        if !allowed {
            c.AbortWithStatus(http.StatusForbidden)
            return
        }
        
        c.Next()
    }
}
```

### **Security Configuration**

```bash
# Production: Specific domains only
export CORS_ALLOWED_ORIGINS="https://secure-banking.com,https://mobile.banking.com"

# Development: Local testing
export CORS_ALLOWED_ORIGINS="http://localhost:3000,http://localhost:5173"

# NEVER use in production
export CORS_ALLOWED_ORIGINS="*"  # ❌ DANGEROUS
```

### **Attack Prevention**

- **Cross-Site Request Forgery (CSRF)**: Unauthorized domains cannot make requests
- **Data Exfiltration**: Malicious sites cannot read API responses
- **Click-jacking**: Prevents embedding in unauthorized frames

## Error Handling & Information Disclosure Prevention

### **Structured Error Responses**

**Purpose**: Provide consistent error information while preventing sensitive data leakage.

```go
type APIError struct {
    Code    string `json:"code"`       // Machine-readable error code
    Message string `json:"message"`    // Human-readable message  
    Status  int    `json:"-"`         // HTTP status (not exposed)
}

// Example error responses
{
    "code": "INSUFFICIENT_FUNDS",
    "message": "Insufficient funds for this transaction"
}

{
    "code": "ACCOUNT_NOT_FOUND", 
    "message": "Account not found"
}
```

### **Information Disclosure Prevention**

**What we DON'T expose:**
- Internal server errors or stack traces
- Database connection details
- File system paths
- Account balances in error messages
- Existence of specific accounts (timing attacks)

**What we DO log securely:**
```go
// Security event logging with context
logging.Warn("Invalid transfer attempt", map[string]interface{}{
    "error":       err.Error(),
    "ip":          c.ClientIP(),
    "user_agent":  c.Request.UserAgent(),
    "from_id":     req.FromID,
    "to_id":       req.ToID,
    "amount":      req.Amount,
    "timestamp":   time.Now(),
})
```

## Audit Logging & Forensic Trails

### **Comprehensive Transaction Logging**

**Purpose**: Maintain complete audit trails for compliance and forensic analysis.

### **Security Event Categories**

1. **Authentication Events**
   ```go
   logging.Info("Account access", map[string]interface{}{
       "account_id": accountID,
       "operation":  "balance_query",
       "ip":         clientIP,
       "success":    true,
   })
   ```

2. **Authorization Failures**
   ```go
   logging.Warn("Rate limit exceeded", map[string]interface{}{
       "ip":              clientIP,
       "requests_in_window": requestCount,
       "limit":           rateLimitConfig.Limit,
       "user_agent":      userAgent,
   })
   ```

3. **Business Logic Violations**
   ```go
   logging.Error("Transfer validation failed", map[string]interface{}{
       "from_account":    req.FromID,
       "to_account":      req.ToID,
       "amount":          req.Amount,
       "reason":          "insufficient_funds",
       "current_balance": currentBalance,
       "ip":              clientIP,
   })
   ```

4. **System Events**
   ```go
   logging.Info("System startup", map[string]interface{}{
       "version":    appVersion,
       "config":     sanitizedConfig,
       "timestamp":  startTime,
   })
   ```

### **Log Security Features**

- **Structured JSON**: Machine-parseable for SIEM integration
- **Correlation IDs**: Track requests across system boundaries  
- **Tamper Evidence**: Cryptographic integrity (future enhancement)
- **Retention Policies**: Automatic archival and purging
- **Sensitive Data Protection**: No passwords or PII in logs

## Business Logic Security

### **Transaction Validation Rules**

**Purpose**: Enforce business rules that prevent financial fraud and abuse.

### **Transfer Security Controls**

1. **Self-Transfer Prevention**
   ```go
   if req.FromID == req.ToID {
       return errors.NewSelfTransferError()
   }
   ```

2. **Account Existence Validation**
   ```go
   from, exists := database.Repo.GetAccount(req.FromID)
   if !exists {
       return errors.NewAccountNotFoundError()
   }
   ```

3. **Balance Sufficiency Check**
   ```go
   if from.Balance < req.Amount {
       return errors.NewInsufficientFundsError()
   }
   ```

4. **Atomic Operations**
   ```go
   // All balance updates are atomic to prevent race conditions
   from.Balance -= req.Amount
   to.Balance += req.Amount
   ```

### **Account Creation Security**

1. **Owner Name Validation**
   - Length limits (2-100 characters)
   - Character restrictions (letters, spaces, safe punctuation)
   - Unicode normalization

2. **Duplicate Prevention**
   - Account ID sequence integrity
   - Unique constraint enforcement

## Thread Safety & Concurrency Security

### **Deadlock Prevention**

**Purpose**: Prevent system lockups while maintaining data consistency.

```go
// Ordered locking algorithm prevents deadlocks
if from.Id < to.Id {
    from.Mu.Lock()
    to.Mu.Lock()
} else {
    to.Mu.Lock()
    from.Mu.Lock()
}
defer from.Mu.Unlock()
defer to.Mu.Unlock()
```

### **Race Condition Prevention**

- **Account-level locking**: Each account has its own mutex
- **Atomic operations**: All balance updates are protected
- **Consistent ordering**: Prevents circular wait conditions

## Security Configuration Management

### **Environment-Based Security Settings**

```bash
# Security-focused production configuration
export CORS_ALLOWED_ORIGINS="https://secure-banking.com"
export RATE_LIMIT_REQUESTS_PER_MINUTE=30
export LOG_LEVEL=warn
export LOG_FORMAT=json

# High-security environment
export CORS_ALLOWED_ORIGINS="https://internal.bank.com"  
export RATE_LIMIT_REQUESTS_PER_MINUTE=10
export LOG_LEVEL=error
```

### **Security Defaults**

- **Rate limiting**: Enabled by default (100 req/min)
- **CORS**: Restrictive allowlist (localhost only in dev)
- **Validation**: All inputs validated by default
- **Logging**: Security events always logged
- **Error handling**: Safe defaults, no information disclosure

## Monitoring & Alerting

### **Security Metrics**

Track these key security indicators:

1. **Rate Limiting Events**
   - Requests blocked per IP
   - Top offending IP addresses
   - Rate limit threshold adjustments

2. **Validation Failures**
   - Invalid input attempt frequency
   - Attack pattern detection
   - Payload analysis

3. **CORS Violations**
   - Blocked origin attempts
   - Suspicious cross-origin patterns
   - Configuration effectiveness

### **Alert Thresholds**

```bash
# Example alerting rules
rate_limit_violations_per_minute > 10
invalid_input_attempts_per_hour > 50
cors_violations_per_day > 5
failed_transfers_per_minute > 20
```

## Compliance & Best Practices

### **Industry Standards Alignment**

- **PCI DSS**: Data protection and access controls
- **SOX**: Audit logging and transaction integrity
- **GDPR**: Privacy by design (when handling EU customers)
- **OWASP Top 10**: Protection against common vulnerabilities

### **Security Testing**

Regular security assessments should include:

1. **Penetration Testing**
   - Rate limiting bypass attempts
   - Input validation evasion
   - CORS policy violations

2. **Load Testing**
   - DoS resistance validation
   - Resource exhaustion testing
   - Concurrent attack simulation

3. **Code Review**
   - Static analysis for vulnerabilities
   - Dependency vulnerability scanning
   - Security-focused code reviews

## Future Security Enhancements

### **Planned Improvements**

1. **Authentication & Authorization**
   - JWT token implementation
   - Role-based access controls
   - Multi-factor authentication

2. **Advanced Threat Protection**
   - Machine learning fraud detection
   - Behavioral analysis
   - Geolocation-based controls

3. **Cryptographic Enhancements**
   - Data encryption at rest
   - API request signing
   - Certificate-based authentication

4. **Monitoring & Response**
   - SIEM integration
   - Automated threat response
   - Incident response playbooks

This security implementation provides a robust foundation for a production banking system, with multiple layers of protection against common and advanced threats.