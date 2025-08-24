package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Collector struct {
	mu             sync.RWMutex
	operations     map[string]*OperationMetrics
	startTime      time.Time
	totalRequests  int64
	totalSuccess   int64
	totalFailures  int64
	latencies      []time.Duration
	errorTypes     map[string]int64
}

type OperationMetrics struct {
	Count     int64
	Success   int64
	Failures  int64
	Latencies []time.Duration
	Errors    map[string]int64
}

type Stats struct {
	TotalRequests     int64
	TotalSuccess      int64
	TotalFailures     int64
	SuccessRate       float64
	RequestsPerSecond float64
	MeanLatency       time.Duration
	MedianLatency     time.Duration
	P50Latency        time.Duration
	P90Latency        time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	MinLatency        time.Duration
	MaxLatency        time.Duration
	StdDevLatency     time.Duration
	OperationStats    map[string]*OperationStats
	ErrorDistribution map[string]int64
	Duration          time.Duration
}

type OperationStats struct {
	Count             int64
	SuccessRate       float64
	MeanLatency       time.Duration
	P99Latency        time.Duration
	ErrorDistribution map[string]int64
}

func NewCollector() *Collector {
	return &Collector{
		operations: make(map[string]*OperationMetrics),
		startTime:  time.Now(),
		errorTypes: make(map[string]int64),
	}
}

func (c *Collector) RecordOperation(opType string, latency time.Duration, success bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.operations[opType]; !exists {
		c.operations[opType] = &OperationMetrics{
			Latencies: make([]time.Duration, 0, 10000),
			Errors:    make(map[string]int64),
		}
	}

	op := c.operations[opType]
	atomic.AddInt64(&op.Count, 1)
	atomic.AddInt64(&c.totalRequests, 1)

	if success {
		atomic.AddInt64(&op.Success, 1)
		atomic.AddInt64(&c.totalSuccess, 1)
	} else {
		atomic.AddInt64(&op.Failures, 1)
		atomic.AddInt64(&c.totalFailures, 1)
		
		if err != nil {
			errStr := err.Error()
			op.Errors[errStr]++
			c.errorTypes[errStr]++
		}
	}

	op.Latencies = append(op.Latencies, latency)
	c.latencies = append(c.latencies, latency)
}

func (c *Collector) GetStats() *Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	duration := time.Since(c.startTime)
	
	stats := &Stats{
		TotalRequests:     atomic.LoadInt64(&c.totalRequests),
		TotalSuccess:      atomic.LoadInt64(&c.totalSuccess),
		TotalFailures:     atomic.LoadInt64(&c.totalFailures),
		Duration:          duration,
		OperationStats:    make(map[string]*OperationStats),
		ErrorDistribution: make(map[string]int64),
	}

	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.TotalSuccess) / float64(stats.TotalRequests)
		stats.RequestsPerSecond = float64(stats.TotalRequests) / duration.Seconds()
	}

	if len(c.latencies) > 0 {
		latenciesCopy := make([]time.Duration, len(c.latencies))
		copy(latenciesCopy, c.latencies)
		sort.Slice(latenciesCopy, func(i, j int) bool {
			return latenciesCopy[i] < latenciesCopy[j]
		})

		stats.MinLatency = latenciesCopy[0]
		stats.MaxLatency = latenciesCopy[len(latenciesCopy)-1]
		stats.MedianLatency = percentile(latenciesCopy, 50)
		stats.P50Latency = percentile(latenciesCopy, 50)
		stats.P90Latency = percentile(latenciesCopy, 90)
		stats.P95Latency = percentile(latenciesCopy, 95)
		stats.P99Latency = percentile(latenciesCopy, 99)
		stats.MeanLatency = mean(latenciesCopy)
		stats.StdDevLatency = stdDev(latenciesCopy, stats.MeanLatency)
	}

	for opType, metrics := range c.operations {
		opStats := &OperationStats{
			Count:             atomic.LoadInt64(&metrics.Count),
			ErrorDistribution: make(map[string]int64),
		}

		if opStats.Count > 0 {
			opStats.SuccessRate = float64(atomic.LoadInt64(&metrics.Success)) / float64(opStats.Count)
			
			if len(metrics.Latencies) > 0 {
				latenciesCopy := make([]time.Duration, len(metrics.Latencies))
				copy(latenciesCopy, metrics.Latencies)
				sort.Slice(latenciesCopy, func(i, j int) bool {
					return latenciesCopy[i] < latenciesCopy[j]
				})
				
				opStats.MeanLatency = mean(latenciesCopy)
				opStats.P99Latency = percentile(latenciesCopy, 99)
			}
		}

		for errType, count := range metrics.Errors {
			opStats.ErrorDistribution[errType] = count
		}

		stats.OperationStats[opType] = opStats
	}

	for errType, count := range c.errorTypes {
		stats.ErrorDistribution[errType] = count
	}

	return stats
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	
	index := int(float64(len(sorted)-1) * p / 100.0)
	return sorted[index]
}

func mean(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}
	
	var sum time.Duration
	for _, v := range values {
		sum += v
	}
	return sum / time.Duration(len(values))
}

func stdDev(values []time.Duration, mean time.Duration) time.Duration {
	if len(values) <= 1 {
		return 0
	}
	
	var sumSquares float64
	meanFloat := float64(mean)
	
	for _, v := range values {
		diff := float64(v) - meanFloat
		sumSquares += diff * diff
	}
	
	variance := sumSquares / float64(len(values)-1)
	return time.Duration(variance)
}