# Performance Optimization Guide

## Current Performance Baseline
- **Throughput**: 4,000-6,500 requests/second
- **P99 Latency**: 190ms
- **Memory Usage**: 482MB
- **CPU Usage**: 100% (saturated)
- **GC Pause Time**: 600ms+
- **Test Configuration**: 500 workers, 0ms think time, 1M total operations

## Identified Bottlenecks

### 1. Double Mutex Locking
- **Issue**: Database uses global `sync.RWMutex` + each account has individual `sync.Mutex`
- **Impact**: High lock contention causing CPU saturation and latency
- **Location**: `src/diplomat/database/inmemory.go` and `src/domain/account.go`

### 2. Excessive Logging
- **Issue**: Every request logs validation warnings, IP addresses, and debug info
- **Impact**: I/O overhead and memory allocations
- **Location**: `src/handlers/*.go`

### 3. Memory Allocations
- **Issue**: No object pooling, frequent allocations for requests/responses
- **Impact**: GC pressure causing 600ms+ pause times
- **Location**: Throughout request handling pipeline

### 4. Inefficient JSON Processing
- **Issue**: `gin.ShouldBindJSON()` allocates new objects on every request
- **Impact**: Memory pressure and CPU overhead
- **Location**: All handler functions

### 5. Prometheus Metrics
- **Issue**: Metrics collected synchronously on every operation
- **Impact**: Additional CPU overhead per request
- **Location**: `src/diplomat/middleware/prometheus.go`

## Optimization Strategies

### Phase 1: Quick Wins (Target: 20K RPS)

#### 1.1 Reduce Logging
```go
// Before
logging.Warn("Invalid owner name", map[string]interface{}{
    "owner": req.Owner,
    "error": err.Error(),
    "ip":    ctx.ClientIP(),
})

// After - Log only critical errors or use debug level
if debugMode {
    logging.Debug("Invalid owner name", ...)
}
```

#### 1.2 Optimize Locking Strategy
```go
// Use RLock for read operations
func GetAccount(id int) (*models.Account, bool) {
    db.mu.RLock()  // Changed from Lock()
    defer db.mu.RUnlock()
    // ...
}
```

#### 1.3 Request Object Pooling
```go
var accountRequestPool = sync.Pool{
    New: func() interface{} {
        return &CreateAccountRequest{}
    },
}

func CreateAccount(ctx *gin.Context) {
    req := accountRequestPool.Get().(*CreateAccountRequest)
    defer accountRequestPool.Put(req)
    // ...
}
```

### Phase 2: Architecture Changes (Target: 100K RPS)

#### 2.1 Lock-Free Account Operations
```go
type Account struct {
    ID      int
    Owner   string
    balance int64  // Use atomic operations
}

func (a *Account) AddBalance(amount int64) {
    atomic.AddInt64(&a.balance, amount)
}

func (a *Account) GetBalance() int64 {
    return atomic.LoadInt64(&a.balance)
}
```

#### 2.2 Replace Map with sync.Map
```go
type InMemory struct {
    accounts sync.Map  // Lock-free concurrent map
    nextID   int64     // Use atomic counter
}

func (db *InMemory) GetAccount(id int) (*models.Account, bool) {
    value, ok := db.accounts.Load(id)
    if !ok {
        return nil, false
    }
    return value.(*models.Account), true
}
```

#### 2.3 Batch Processing
```go
type BatchProcessor struct {
    operations chan Operation
    batchSize  int
    interval   time.Duration
}

func (bp *BatchProcessor) ProcessBatch(ops []Operation) {
    // Group by account ID to minimize lock acquisitions
    grouped := make(map[int][]Operation)
    for _, op := range ops {
        grouped[op.AccountID] = append(grouped[op.AccountID], op)
    }
    
    // Process each account's operations in one lock
    for accountID, accountOps := range grouped {
        processAccountOperations(accountID, accountOps)
    }
}
```

### Phase 3: Advanced Optimizations (Target: 1M RPS)

#### 3.1 Sharded Database
```go
type ShardedDB struct {
    shards   []*InMemory
    numShards int
}

func (s *ShardedDB) getShard(accountID int) *InMemory {
    return s.shards[accountID % s.numShards]
}
```

#### 3.2 Zero-Allocation JSON
```go
// Use libraries like easyjson or jsoniter
//go:generate easyjson -all models.go

func parseRequest(data []byte) (*Request, error) {
    req := requestPool.Get().(*Request)
    err := req.UnmarshalJSON(data)
    return req, err
}
```

#### 3.3 Custom HTTP Server
```go
// Replace Gin with fasthttp for better performance
import "github.com/valyala/fasthttp"

func handleRequest(ctx *fasthttp.RequestCtx) {
    // 10x faster than standard net/http
}
```

## Implementation Priority

1. **Immediate** (1-2 hours, 5x improvement)
   - Remove excessive logging
   - Fix lock types (RLock vs Lock)
   - Add request pooling

2. **Short-term** (1 day, 20x improvement)
   - Implement atomic operations for balances
   - Replace map with sync.Map
   - Optimize JSON processing

3. **Medium-term** (1 week, 100x improvement)
   - Implement sharded database
   - Add batch processing
   - Switch to fasthttp

## Testing Strategy

### Benchmark Each Change
```bash
# Before optimization
go test -bench=. -benchmem -cpuprofile=cpu.prof

# After optimization
go test -bench=. -benchmem -cpuprofile=cpu_optimized.prof

# Compare profiles
go tool pprof -diff_base=cpu.prof cpu_optimized.prof
```

### Load Testing Scenarios
1. **Baseline**: Current implementation
2. **Step 1**: With logging removed
3. **Step 2**: With optimized locking
4. **Step 3**: With atomic operations
5. **Final**: All optimizations combined

## Monitoring Metrics

- **Request Rate**: Track RPS improvements
- **Latency**: P50, P95, P99 percentiles
- **GC Metrics**: Pause time, frequency, heap size
- **CPU Profile**: Identify hot paths
- **Memory Profile**: Track allocation rate
- **Lock Contention**: Monitor mutex wait times

## Expected Results

| Phase | Optimization | Expected RPS | P99 Latency | GC Pause |
|-------|-------------|--------------|-------------|----------|
| Baseline | Current | 6.5K | 190ms | 600ms |
| Phase 1 | Quick Wins | 20-30K | 50ms | 100ms |
| Phase 2 | Architecture | 100-150K | 10ms | 20ms |
| Phase 3 | Advanced | 500K-1M | 2ms | 5ms |

## Next Steps

1. Create feature branch for optimizations
2. Implement Phase 1 changes
3. Run performance tests after each change
4. Document actual vs expected improvements
5. Proceed to next phase based on results