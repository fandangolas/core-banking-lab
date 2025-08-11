# Observability & Monitoring

**Comprehensive monitoring, logging, and metrics for production-ready banking operations.**

## Overview

The banking API implements **full observability stack** with structured logging, metrics collection, real-time events, and health monitoring—essential for production financial systems.

## Structured Logging

### JSON Logging for Production
```go
type Logger struct {
    *slog.Logger
}

func Info(message string, fields map[string]interface{}) {
    logger.With(convertFields(fields)...).Info(message)
}

// Example output:
{
    "time": "2025-08-10T02:54:47Z",
    "level": "INFO",
    "msg": "Transfer completed",
    "from_account": 1,
    "to_account": 2,
    "amount": 5000,
    "request_id": "req_123",
    "duration_ms": 1.2
}
```

### Context-Rich Logging
```go
// Request-level context
logging.Info("Account balance retrieved", map[string]interface{}{
    "account_id": accountID,
    "balance": balance,
    "client_ip": clientIP,
    "user_agent": userAgent,
    "response_time_ms": duration,
})

// Error logging with context
logging.Error("Transfer failed", map[string]interface{}{
    "error": err.Error(),
    "from_account": fromID,
    "to_account": toID,
    "amount": amount,
    "reason": "insufficient_funds",
})
```

**Benefits:**
- **Searchable**: Query logs by account ID, error type, etc.
- **Structured**: Machine-readable for log aggregation
- **Contextual**: Every log entry has relevant business context
- **Audit Trail**: Complete transaction history for compliance

## Real-Time Metrics

### Performance Metrics
```go
type Metrics struct {
    TotalRequests     int64
    TotalErrors       int64
    AverageLatency    float64
    EndpointStats     map[string]EndpointMetrics
    SystemInfo        SystemMetrics
}

type EndpointMetrics struct {
    Count           int64   `json:"count"`
    AverageLatency  float64 `json:"avg_duration_ms"`
    ErrorRate       float64 `json:"error_rate"`
}
```

### System Health Monitoring
```go
func GetSystemMetrics() SystemMetrics {
    return SystemMetrics{
        UptimeSeconds: int64(time.Since(startTime).Seconds()),
        Goroutines:    runtime.NumGoroutine(),
        MemoryMB:      getMemoryUsage(),
        CPUPercent:    getCPUUsage(),
    }
}
```

### API Endpoint: `/metrics`
```bash
curl http://localhost:8080/metrics

# Response:
{
    "endpoints": {
        "POST /accounts": {
            "count": 150,
            "avg_duration_ms": 0.3,
            "error_rate": 0.02
        },
        "POST /accounts/transfer": {
            "count": 445, 
            "avg_duration_ms": 1.2,
            "error_rate": 0.001
        }
    },
    "system": {
        "uptime_seconds": 3600,
        "goroutines": 15,
        "memory_mb": 45,
        "cpu_percent": 12.5
    }
}
```

## Real-Time Event Streaming

### WebSocket Event Publishing
```go
type EventBroker struct {
    subscribers []chan TransactionEvent
    mutex       sync.RWMutex
}

// Non-blocking event publishing
func (eb *EventBroker) Publish(event TransactionEvent) {
    eb.mutex.RLock()
    defer eb.mutex.RUnlock()
    
    for _, subscriber := range eb.subscribers {
        select {
        case subscriber <- event:
            // Event delivered
        default:
            // Skip slow subscribers (no blocking)
        }
    }
}
```

### Live Transaction Events
```json
{
    "type": "transfer",
    "from_id": 1,
    "to_id": 2,
    "amount": 5000,
    "from_balance": 10000,
    "to_balance": 5000,
    "timestamp": "2025-08-10T02:54:47Z",
    "request_id": "req_456"
}
```

### Dashboard Integration
```javascript
const ws = new WebSocket('ws://localhost:8080/events');

ws.onmessage = function(event) {
    const transaction = JSON.parse(event.data);
    updateDashboard(transaction);
};

// Real-time balance updates
// Live transaction feed  
// System health indicators
```

## Health Monitoring

### Health Check Endpoint
```bash
curl http://localhost:8080/health

# Response:
{
    "status": "healthy",
    "timestamp": "2025-08-10T02:54:47Z",
    "version": "1.0.0",
    "database": "connected",
    "memory_usage": "45MB",
    "goroutines": 15
}
```

### Kubernetes Health Probes
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health  
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Request Tracing

### Request ID Tracking
```go
// Middleware adds unique ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        requestID := generateRequestID()
        c.Header("X-Request-ID", requestID)
        c.Set("request_id", requestID)
        
        logging.Info("Request started", map[string]interface{}{
            "request_id": requestID,
            "method": c.Request.Method,
            "path": c.Request.URL.Path,
            "client_ip": c.ClientIP(),
        })
        
        c.Next()
    })
}
```

### Full Request Lifecycle Logging
```json
// Request start
{"level": "INFO", "msg": "Request started", "request_id": "req_123", "path": "/accounts/transfer"}

// Business logic
{"level": "INFO", "msg": "Transfer initiated", "request_id": "req_123", "from": 1, "to": 2}

// Request completion
{"level": "INFO", "msg": "Request completed", "request_id": "req_123", "status": 200, "duration_ms": 1.2}
```

## Production Monitoring Setup

### Docker Compose with Monitoring Stack
```yaml
version: '3.8'
services:
  api:
    build: .
    environment:
      - LOG_LEVEL=info
      - LOG_FORMAT=json
    
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

### Key Metrics to Monitor

**Application Metrics:**
- Request rate (requests/second)
- Response time percentiles (P50, P95, P99)  
- Error rate percentage
- Concurrent goroutines count

**Business Metrics:**
- Accounts created per hour
- Total transaction volume
- Transfer success rate
- Balance query frequency

**System Metrics:**
- CPU utilization
- Memory usage
- Database connection pool status
- WebSocket connections

## Error Tracking & Alerting

### Structured Error Handling
```go
type APIError struct {
    Code      string    `json:"code"`
    Message   string    `json:"message"`
    RequestID string    `json:"request_id,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

// Log error with full context
logging.Error("Transfer validation failed", map[string]interface{}{
    "error_code": "INSUFFICIENT_FUNDS",
    "account_id": fromID,
    "attempted_amount": amount,
    "current_balance": balance,
    "request_id": requestID,
    "client_ip": clientIP,
})
```

### Alert Conditions
```go
// Monitor these conditions for production alerts:

// High error rate
if errorRate > 5.0 {
    alert("High error rate: " + errorRate + "%")
}

// High response time  
if p95Latency > 100 {
    alert("High latency: P95 = " + p95Latency + "ms")
}

// Memory leak detection
if goroutineCount > 1000 {
    alert("Possible goroutine leak: " + goroutineCount)
}
```

## Log Aggregation

### ELK Stack Integration
```yaml
# Filebeat configuration
filebeat.inputs:
- type: container
  paths:
    - '/var/lib/docker/containers/*/*.log'
  
output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  
processors:
- add_kubernetes_metadata: ~
- decode_json_fields:
    fields: ["message"]
    target: ""
```

### Log Queries for Operations
```bash
# Find all errors for specific account
GET /logs/_search
{
  "query": {
    "bool": {
      "must": [
        {"term": {"level": "ERROR"}},
        {"term": {"account_id": 1}}
      ]
    }
  }
}

# Transaction audit trail
GET /logs/_search  
{
  "query": {
    "match": {"request_id": "req_123"}
  },
  "sort": [{"timestamp": {"order": "asc"}}]
}
```

## Performance Profiling

### Built-in Profiling Endpoints
```go
import _ "net/http/pprof"

// Enable profiling in development
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

### Analysis Commands
```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profiling  
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine analysis
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Key Observability Features

✅ **Structured Logging**: JSON format with business context  
✅ **Real-time Metrics**: Live performance and system health data  
✅ **Event Streaming**: WebSocket-based transaction notifications  
✅ **Request Tracing**: Full request lifecycle with unique IDs  
✅ **Health Monitoring**: Kubernetes-ready health checks  
✅ **Error Tracking**: Comprehensive error context and audit trails  

This observability implementation provides complete visibility into system behavior, enabling rapid troubleshooting, performance optimization, and compliance reporting for production banking operations.