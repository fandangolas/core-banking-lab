package testenv

import (
	"bank-api/internal/config"
	"bank-api/internal/infrastructure/database"
	"bank-api/internal/infrastructure/events"
	"bank-api/internal/infrastructure/messaging"
	"bank-api/internal/pkg/logging"
	"log"

	"github.com/gin-gonic/gin"
)

// TestContainer is a lightweight version of the components.Container for testing
type TestContainer struct {
	Config         *config.Config
	Database       database.Repository
	EventBroker    *events.Broker
	EventPublisher *messaging.EventCapture
	Router         *gin.Engine
}

// NewTestContainer creates a test container with minimal setup
// Note: This function expects the database to be already initialized via SetupIntegrationTest(t)
// Call SetupIntegrationTest(t) before calling this function to ensure database.Repo is set
func NewTestContainer() *TestContainer {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize minimal config for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Host: "localhost",
		},
		Database: config.DatabaseConfig{
			Type: "postgres",
		},
		Logging: config.LoggingConfig{
			Level:  "error",
			Format: "json",
		},
		Environment: "test",
		CORS: config.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"*"},
		},
	}

	// Initialize logging in test mode
	logging.Init(cfg)

	// Use the already-initialized database repository from SetupIntegrationTest
	// This avoids creating duplicate connections
	if database.Repo == nil {
		log.Fatal("Database repository not initialized. Call SetupIntegrationTest(t) before NewTestContainer()")
	}
	db := database.Repo

	// Get the singleton event broker
	eventBroker := events.GetBroker()

	// Create event capture for testing
	eventPublisher := messaging.NewEventCapture()

	// Create router with event publisher
	router := SetupTestRouterWithEventPublisher(eventPublisher)

	return &TestContainer{
		Config:         cfg,
		Database:       db,
		EventBroker:    eventBroker,
		EventPublisher: eventPublisher,
		Router:         router,
	}
}

// Reset clears all data in the test container
func (tc *TestContainer) Reset() {
	if tc.Database != nil {
		tc.Database.Reset()
	}
	if tc.EventPublisher != nil {
		tc.EventPublisher.Reset()
	}
}

// GetRouter returns the test router
func (tc *TestContainer) GetRouter() *gin.Engine {
	return tc.Router
}

// GetDatabase returns the test database
func (tc *TestContainer) GetDatabase() database.Repository {
	return tc.Database
}

// GetEventBroker returns the test event broker
func (tc *TestContainer) GetEventBroker() *events.Broker {
	return tc.EventBroker
}

// GetEventPublisher returns the test event publisher (EventCapture)
func (tc *TestContainer) GetEventPublisher() *messaging.EventCapture {
	return tc.EventPublisher
}
