package metrics

import (
	"sync"
	"time"
)

// RequestMetric stores basic information about an HTTP request.
type RequestMetric struct {
	Endpoint string
	Status   int
	Duration time.Duration
}

var (
	mu         sync.Mutex
	metricList []RequestMetric
)

// Record adds a new metric entry in a thread-safe way.
func Record(endpoint string, status int, duration time.Duration) {
	mu.Lock()
	metricList = append(metricList, RequestMetric{Endpoint: endpoint, Status: status, Duration: duration})
	mu.Unlock()
}

// List returns a copy of the collected metrics.
func List() []RequestMetric {
	mu.Lock()
	defer mu.Unlock()
	copied := make([]RequestMetric, len(metricList))
	copy(copied, metricList)
	return copied
}
