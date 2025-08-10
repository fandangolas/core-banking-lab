# Observability & Monitoring

## Overview

This document covers the comprehensive observability implementation in Core Banking Lab, including structured logging, metrics collection, and monitoring strategies for production banking systems.

## Structured Logging

### **JSON-Based Logging System**

**Purpose**: Provide machine-readable, searchable logs for monitoring and debugging.

```go
type LogEntry struct {
    Timestamp string                 `json:"timestamp"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Example log output
{
    "timestamp": "2025-08-10T02:54:47Z",
    "level": "INFO",
    "message": "Account created successfully",
    "fields": {
        "account_id": 1,
        "owner": "Alice",
        "ip": "127.0.0.1",
        "request_id": "req-abc123"
    }
}
```

### **Log Levels & Categories**

**DEBUG**: Development diagnostics
```go
logging.Debug("Processing transfer request", map[string]interface{}{
    "from_id": req.FromID,
    "to_id": req.ToID,
    "amount": req.Amount,
})
```

**INFO**: Business events and system status
```go
logging.Info("Transfer completed", map[string]interface{}{
    "transaction_id": txnID,
    "duration_ms": elapsed.Milliseconds(),
    "success": true,
})
```

**WARN**: Security events and recoverable errors
```go
logging.Warn("Rate limit approached", map[string]interface{}{
    "ip": clientIP,
    "requests_count": currentCount,
    "limit": rateLimitConfig.Limit,
})
```

**ERROR**: System errors requiring attention
```go
logging.Error("Database connection failed", map[string]interface{}{
    "error": err.Error(),
    "retry_attempt": retryCount,
    "max_retries": maxRetries,
})
```

## Request Tracing

### **Correlation ID Implementation**

Track requests across system boundaries:

```go
// Middleware to inject correlation IDs
func RequestTracing() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        correlationID := param.Keys["correlation_id"]
        
        logEntry := map[string]interface{}{
            "correlation_id": correlationID,
            "method": param.Method,
            "path": param.Path,
            "status": param.StatusCode,
            "duration_ms": param.Latency.Milliseconds(),
            "ip": param.ClientIP,
            "user_agent": param.Request.UserAgent(),
        }
        
        jsonLog, _ := json.Marshal(logEntry)
        return string(jsonLog) + "\n"
    })
}
```

### **End-to-End Request Flow**

```
HTTP Request → [correlation_id: req-abc123]
  ↓ Rate Limiter → [correlation_id: req-abc123, component: rate_limiter]
  ↓ Validation → [correlation_id: req-abc123, component: validator]  
  ↓ Business Logic → [correlation_id: req-abc123, component: domain]
  ↓ Repository → [correlation_id: req-abc123, component: database]
  ↓ Event Publish → [correlation_id: req-abc123, component: events]
HTTP Response → [correlation_id: req-abc123, status: 200, duration: 25ms]
```

## Metrics Collection

### **Application Metrics**

**Endpoint Performance Metrics**:
```go
type Metrics struct {
    RequestCount    map[string]int     // Requests per endpoint
    RequestLatency  map[string][]time.Duration  // Response times
    ErrorCount      map[string]int     // Errors per endpoint
    ActiveRequests  int                // Current concurrent requests
}

// Middleware for automatic metrics collection
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        endpoint := c.Request.Method + " " + c.FullPath()
        
        // Record metrics
        recordRequestMetrics(endpoint, duration, c.Writer.Status())
    }
}
```

**Business Logic Metrics**:
```go
// Account operation metrics
var (
    AccountsCreated     int64
    TransfersCompleted  int64
    TransfersFailed     int64
    TotalVolumeTransferred int64
    AverageTransferAmount  float64
)

// Update metrics in business logic
func RecordTransfer(amount int, success bool) {
    if success {
        atomic.AddInt64(&TransfersCompleted, 1)
        atomic.AddInt64(&TotalVolumeTransferred, int64(amount))
    } else {
        atomic.AddInt64(&TransfersFailed, 1)
    }
}
```

### **System Health Metrics**

**Resource Utilization**:
```go
type SystemMetrics struct {
    CPUUsage        float64
    MemoryUsage     int64
    GoroutineCount  int
    DatabaseConns   int
    ActiveSessions  int
}

// Health check endpoint
func HealthCheck(c *gin.Context) {
    metrics := SystemMetrics{
        CPUUsage:       getCPUUsage(),
        MemoryUsage:    getMemoryUsage(),
        GoroutineCount: runtime.NumGoroutine(),
    }
    
    status := "healthy"
    if metrics.CPUUsage > 80 || metrics.MemoryUsage > 1024*1024*1024 {
        status = "degraded"
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status": status,
        "metrics": metrics,
        "timestamp": time.Now(),
    })
}
```

## Real-Time Event Streaming

### **WebSocket-Based Dashboard Updates**

**Event Broker Implementation**:
```go
type EventBroker struct {
    subscribers []chan models.TransactionEvent
    mutex       sync.RWMutex
}

func (eb *EventBroker) Publish(event models.TransactionEvent) {
    eb.mutex.RLock()
    defer eb.mutex.RUnlock()
    
    for _, subscriber := range eb.subscribers {
        select {
        case subscriber <- event:
            // Event delivered
        default:
            // Skip slow subscribers (non-blocking)
        }
    }
}

// WebSocket handler for real-time updates
func EventStream(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    eventChan := make(chan models.TransactionEvent, 100)
    events.BrokerInstance.Subscribe(eventChan)
    
    for event := range eventChan {
        if err := conn.WriteJSON(event); err != nil {
            break
        }
    }
}
```

**Event Types Published**:
```go
type TransactionEvent struct {
    Type        string    `json:"type"`        // "transfer", "deposit", "withdraw"
    AccountID   int       `json:"account_id"`  // Primary account
    FromID      int       `json:"from_id"`     // Transfer source
    ToID        int       `json:"to_id"`       // Transfer destination  
    Amount      int       `json:"amount"`      // Transaction amount
    Balance     int       `json:"balance"`     // Updated balance
    Timestamp   time.Time `json:"timestamp"`   // Event time
}
```

## Performance Monitoring

### **Response Time Analysis**

**Latency Tracking**:
```go
// Track P50, P95, P99 latencies per endpoint
type LatencyTracker struct {
    mutex    sync.RWMutex
    samples  []time.Duration
    endpoint string
}

func (lt *LatencyTracker) Record(duration time.Duration) {
    lt.mutex.Lock()
    defer lt.mutex.Unlock()
    
    lt.samples = append(lt.samples, duration)
    
    // Keep only last 1000 samples for rolling window
    if len(lt.samples) > 1000 {
        lt.samples = lt.samples[1:]
    }
}

func (lt *LatencyTracker) GetPercentiles() map[string]time.Duration {
    lt.mutex.RLock()
    defer lt.mutex.RUnlock()
    
    if len(lt.samples) == 0 {
        return map[string]time.Duration{}
    }
    
    sorted := make([]time.Duration, len(lt.samples))
    copy(sorted, lt.samples)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i] < sorted[j]
    })
    
    return map[string]time.Duration{
        "p50": sorted[len(sorted)*50/100],
        "p95": sorted[len(sorted)*95/100], 
        "p99": sorted[len(sorted)*99/100],
    }
}
```

### **Throughput Metrics**

**Requests Per Second**:
```go
type ThroughputTracker struct {
    requestsPerSecond map[int64]int
    mutex             sync.RWMutex
}

func (tt *ThroughputTracker) RecordRequest() {
    now := time.Now().Unix()
    
    tt.mutex.Lock()
    defer tt.mutex.Unlock()
    
    tt.requestsPerSecond[now]++
    
    // Cleanup old data
    cutoff := now - 300 // Keep 5 minutes of data
    for timestamp := range tt.requestsPerSecond {
        if timestamp < cutoff {
            delete(tt.requestsPerSecond, timestamp)
        }
    }
}
```

## Error Tracking & Alerting

### **Error Classification**

**Business Logic Errors** (Expected):
```go
// These are normal business rule violations
logging.Info("Transfer rejected", map[string]interface{}{
    "reason": "insufficient_funds",
    "account_id": accountID,
    "requested_amount": amount,
    "current_balance": balance,
})
```

**System Errors** (Unexpected):
```go
// These require immediate attention
logging.Error("Database connection lost", map[string]interface{}{
    "error": err.Error(),
    "database_host": dbConfig.Host,
    "retry_count": retryCount,
    "alert": true,  // Trigger alerting
})
```

### **Alert Conditions**

**Critical Alerts**:
- Error rate > 5% over 5 minutes
- Response time P95 > 1000ms over 2 minutes  
- Rate limit violations > 100/minute
- Database connection failures
- System resource exhaustion (CPU > 90%, Memory > 90%)

**Warning Alerts**:
- Error rate > 1% over 10 minutes
- Response time P95 > 500ms over 5 minutes
- High concurrent request count
- Unusual transaction patterns

## Dashboard & Visualization

### **Key Performance Indicators**

**System Health Dashboard**:
- Request throughput (req/sec)
- Average response time
- Error rate percentage
- Active user sessions
- System resource utilization

**Business Metrics Dashboard**:
- Total accounts created
- Transaction volume (money transferred)
- Most active accounts
- Transaction success rate
- Average transaction size

### **Real-Time Monitoring**

**Live Transaction Feed**:
```javascript
// Dashboard WebSocket client
const eventSource = new WebSocket('ws://localhost:8080/events');

eventSource.onmessage = function(event) {
    const transaction = JSON.parse(event.data);
    
    // Update dashboard in real-time
    updateAccountBalance(transaction.account_id, transaction.balance);
    addTransactionToFeed(transaction);
    updateMetrics(transaction);
};
```

**Performance Charts**:
- Response time trends
- Request volume over time
- Error rate trends
- Account creation growth
- Transaction volume trends

## Log Management

### **Log Rotation & Retention**

**Configuration**:
```bash
# Production logging configuration
export LOG_LEVEL=info
export LOG_FORMAT=json
export LOG_MAX_SIZE=100MB
export LOG_MAX_AGE=30
export LOG_MAX_BACKUPS=10
```

**Log Processing Pipeline**:
```
Application Logs → Log Aggregator → Search Index → Dashboard
     ↓                    ↓              ↓           ↓
  JSON Format    →    Parse/Filter  →   Store    →  Query
```

### **Security Event Monitoring**

**SIEM Integration Ready**:
```go
// Security events with standardized format for SIEM
func logSecurityEvent(eventType string, details map[string]interface{}) {
    securityEvent := map[string]interface{}{
        "event_type": eventType,
        "severity": "high",
        "source_ip": details["ip"],
        "user_agent": details["user_agent"],
        "timestamp": time.Now(),
        "details": details,
    }
    
    logging.Warn("Security event", securityEvent)
    
    // Could also send to dedicated security log stream
    sendToSecurityTeam(securityEvent)
}
```

## Monitoring Best Practices

### **Production Readiness Checklist**

- ✅ Structured JSON logging with correlation IDs
- ✅ Error classification and alerting thresholds  
- ✅ Performance metrics collection (latency, throughput)
- ✅ Real-time event streaming for dashboards
- ✅ Health check endpoints for load balancers
- ✅ Resource utilization monitoring
- ✅ Security event logging and alerting
- ✅ Log retention and rotation policies

### **Operational Workflows**

**Incident Response**:
1. Alert triggers from metrics threshold
2. Correlation ID lookup for affected requests
3. Log analysis to identify root cause
4. Performance metrics to assess impact
5. Real-time monitoring during resolution

**Capacity Planning**:
1. Trend analysis from historical metrics
2. Load testing with performance monitoring
3. Resource utilization forecasting
4. Scaling decision support from data

This observability implementation provides comprehensive visibility into system behavior, performance, and security for production banking operations.