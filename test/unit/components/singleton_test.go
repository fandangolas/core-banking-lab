package components

import (
	"bank-api/internal/pkg/components"
	"bank-api/internal/infrastructure/database"
	"bank-api/internal/infrastructure/events"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDatabaseSingleton verifies database repository is a true singleton
func TestDatabaseSingleton(t *testing.T) {
	// Reset any previous state
	database.Init()
	repo1 := database.Repo

	// Call Init again
	database.Init()
	repo2 := database.Repo

	// Should be the same instance (singleton behavior)
	assert.Same(t, repo1, repo2, "Database repository should be singleton")
}

// TestEventBrokerSingleton verifies event broker is a true singleton
func TestEventBrokerSingleton(t *testing.T) {
	broker1 := events.GetBroker()
	broker2 := events.GetBroker()

	// Should be the same instance
	assert.Same(t, broker1, broker2, "Event broker should be singleton")
}

// TestComponentsContainerSingleton verifies components container is singleton
func TestComponentsContainerSingleton(t *testing.T) {
	container1, err1 := components.GetInstance()
	container2, err2 := components.GetInstance()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Same(t, container1, container2, "Components container should be singleton")
}

// TestConcurrentSingletonAccess verifies thread-safety of singletons
func TestConcurrentSingletonAccess(t *testing.T) {
	const numGoroutines = 100
	var wg sync.WaitGroup

	// Collect all database instances created concurrently
	dbInstances := make([]interface{}, numGoroutines)
	brokerInstances := make([]interface{}, numGoroutines)
	containerInstances := make([]interface{}, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			// Initialize database
			database.Init()
			dbInstances[index] = database.Repo

			// Get event broker
			brokerInstances[index] = events.GetBroker()

			// Get container
			container, _ := components.GetInstance()
			containerInstances[index] = container
		}(i)
	}

	wg.Wait()

	// Verify all instances are the same (true singleton behavior)
	firstDB := dbInstances[0]
	firstBroker := brokerInstances[0]
	firstContainer := containerInstances[0]

	for i := 1; i < numGoroutines; i++ {
		assert.Same(t, firstDB, dbInstances[i],
			"Database instance %d should be same as first", i)
		assert.Same(t, firstBroker, brokerInstances[i],
			"Broker instance %d should be same as first", i)
		assert.Same(t, firstContainer, containerInstances[i],
			"Container instance %d should be same as first", i)
	}
}

// TestSingletonInitializationOrder verifies initialization happens in correct order
func TestSingletonInitializationOrder(t *testing.T) {
	// Get container (this should initialize everything)
	container, err := components.GetInstance()
	assert.NoError(t, err)

	// Verify all singletons are properly initialized
	assert.NotNil(t, container.Database, "Database should be initialized")
	assert.NotNil(t, container.EventBroker, "Event broker should be initialized")
	assert.NotNil(t, container.Config, "Config should be initialized")

	// Verify they match the global singletons
	assert.Same(t, database.Repo, container.Database,
		"Container database should match global singleton")
	assert.Same(t, events.GetBroker(), container.EventBroker,
		"Container event broker should match global singleton")
}
