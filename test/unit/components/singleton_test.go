package components

import (
	"bank-api/internal/infrastructure/events"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEventBrokerSingleton verifies event broker is a true singleton
func TestEventBrokerSingleton(t *testing.T) {
	broker1 := events.GetBroker()
	broker2 := events.GetBroker()

	// Should be the same instance
	assert.Same(t, broker1, broker2, "Event broker should be singleton")
}

// TestConcurrentEventBrokerAccess verifies thread-safety of event broker singleton
func TestConcurrentEventBrokerAccess(t *testing.T) {
	const numGoroutines = 100
	var wg sync.WaitGroup

	// Collect all instances created concurrently
	brokerInstances := make([]interface{}, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			brokerInstances[index] = events.GetBroker()
		}(i)
	}

	wg.Wait()

	// Verify all instances are the same (true singleton behavior)
	firstBroker := brokerInstances[0]

	for i := 1; i < numGoroutines; i++ {
		assert.Same(t, firstBroker, brokerInstances[i],
			"Broker instance %d should be same as first", i)
	}
}
