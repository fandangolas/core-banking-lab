package metrics

import (
	"sync"
	"time"
)

// RequestMetric stores basic information about an HTTP request.
type RequestMetric struct {
	Endpoint string        `json:"endpoint"`
	Status   int           `json:"status"`
	Duration time.Duration `json:"duration"`
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
