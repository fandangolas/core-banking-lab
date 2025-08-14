# Performance Analysis: Dashboard Stress Testing Issues

**Date:** 2025-08-13  
**Issue:** Dashboard becomes extremely slow when testing with large amounts (100,000 requests)

## Executive Summary

Investigation revealed that using the React dashboard for high-volume stress testing creates fundamental performance bottlenecks. The issue is not with the API performance (which is well-optimized) but with the frontend architecture handling massive concurrent operations.

## Root Cause Analysis

### 1. 🚨 JavaScript Event Loop Overload (Critical)

**Problem:** The `runBatchedSimulation` function in `simulator.js` creates catastrophic inefficiencies:
- **100,000 individual `setTimeout` calls** overwhelming the JavaScript event loop
- **100,000 Promise objects** held simultaneously in memory
- Precise timing intervals create scheduling chaos in the browser

**Code Location:** `/dev/dashboard/src/simulator.js:175-217`

```javascript
// Problematic pattern - creates 100k setTimeout calls
for (let i = 0; i < n; i++) {
  const requestPromise = new Promise((resolve) => {
    setTimeout(async () => {
      // Each request gets its own timer
    }, i * interval);
  });
  blockPromises.push(requestPromise);
}
```

### 2. 🔄 Excessive DOM Re-renders (Critical)

**Problem:** Every request completion triggers React state updates:
- **100,000 individual `setProgress` calls** causing constant DOM re-renders
- No debouncing or batching of UI updates
- Browser struggles with continuous state changes

**Code Location:** `/dev/dashboard/src/App.jsx:339,352,364`

```javascript
// Triggers on every single request completion
setProgress(prev => ({ ...prev, current: prev.current + 1 }));
```

### 3. 💾 Memory Consumption (High)

**Issues:**
- All Promise objects for pending requests accumulate in memory
- setTimeout handles pile up in the event loop queue
- React state updates create intermediate objects constantly

### 4. ✅ API Performance (Actually Good)

**Analysis shows the Go API is well-optimized:**
- Simple in-memory operations with proper mutex locking
- Minimal processing overhead per request
- Resource limits are appropriate (512Mi RAM, 500m CPU)

**Code Evidence:** Clean domain logic in `/src/domain/account.go`:
```go
func withAccountLock(acc *models.Account, fn func()) {
    acc.Mu.Lock()
    defer acc.Mu.Unlock()
    fn()
}
```

### 5. ✅ Infrastructure (Acceptable)

**Kubernetes resource constraints are appropriate for t4g.small:**
- Memory: 128Mi requests, 512Mi limits
- CPU: 100m requests, 500m limits
- Network overhead is minimal for individual requests

## Performance Metrics Impact

| Component | Impact Level | Reason |
|-----------|-------------|---------|
| JavaScript Event Loop | **Critical** | 100k setTimeout calls |
| DOM Rendering | **Critical** | 100k state updates |
| Memory Usage | **High** | Promise accumulation |
| API Processing | **Minimal** | Well-optimized Go code |
| Network | **Minimal** | Simple HTTP requests |
| Infrastructure | **Low** | Adequate resources |

## Recommendations

### Immediate Actions

1. **Stop using dashboard for high-volume testing**
   - Dashboard is designed for monitoring and demos
   - Browser architecture fundamentally unsuited for load generation

2. **Use proper load testing tools:**
   - **curl/wget**: Simple command-line testing
   - **hey/ab/wrk**: Dedicated HTTP load testers
   - **artillery/k6**: Professional load testing platforms
   - **Custom Go tool**: Could build dedicated CLI load tester

### Frontend Improvements (If Still Needed)

1. **Debounce UI updates:**
   ```javascript
   // Update progress every 100ms instead of every request
   const debouncedSetProgress = useMemo(
     () => debounce(setProgress, 100),
     []
   );
   ```

2. **Use Web Workers:**
   - Move simulation logic off main thread
   - Prevent UI blocking during load tests

3. **Batch operations:**
   - Group requests into larger batches
   - Reduce total number of setTimeout calls

### Alternative Testing Approaches

#### Option 1: Simple curl commands
```bash
# Test 1000 deposits
for i in {1..1000}; do
  curl -X POST http://api:8080/accounts/1/deposit \
    -H "Content-Type: application/json" \
    -d '{"amount": 100}' &
done
wait
```

#### Option 2: Go-based load tester
Create dedicated CLI tool that bypasses browser limitations entirely.

#### Option 3: Professional tools
```bash
# Using 'hey' tool
hey -n 100000 -c 100 -m POST \
  -H "Content-Type: application/json" \
  -d '{"amount": 100}' \
  http://api:8080/accounts/1/deposit
```

## Conclusion

The performance issue is **architectural**, not implementational. The React dashboard performs well for its intended use case (monitoring, demos, moderate testing), but browsers are fundamentally not designed for high-volume load generation.

**Key Takeaway:** Use the right tool for the job - dashboards for monitoring, dedicated tools for stress testing.

## Files Investigated

- `/dev/dashboard/src/simulator.js` - Simulation logic
- `/dev/dashboard/src/App.jsx` - React state management
- `/src/handlers/deposit.go` - API endpoint implementation
- `/src/domain/account.go` - Core business logic
- `/k8s/01-api-deployment.yaml` - Resource constraints