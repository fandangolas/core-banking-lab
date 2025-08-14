package testenv

import (
	"bank-api/src/config"
	"bank-api/src/diplomat/database"
	"bank-api/src/diplomat/events"
	"bank-api/src/logging"

	"github.com/gin-gonic/gin"
)

// TestContainer is a lightweight version of the components.Container for testing
type TestContainer struct {
	Config      *config.Config
	Database    database.Repository
	EventBroker *events.Broker
	Router      *gin.Engine
}

// NewTestContainer creates a test container with minimal setup
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
			Type: "inmemory",
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

	// Initialize database
	database.Init()
	db := database.Repo

	// Initialize event broker
	eventBroker := events.NewBroker()
	events.BrokerInstance = eventBroker

	// Create router with middleware and routes
	router := SetupTestRouter()

	return &TestContainer{
		Config:      cfg,
		Database:    db,
		EventBroker: eventBroker,
		Router:      router,
	}
}

// Reset clears all data in the test container
func (tc *TestContainer) Reset() {
	if tc.Database != nil {
		tc.Database.Reset()
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