package testenv

import (
	"bank-api/internal/config"
	"bank-api/internal/infrastructure/database"
	"bank-api/internal/infrastructure/database/postgres"
	"bank-api/internal/infrastructure/events"
	"bank-api/internal/pkg/logging"
	"log"

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

	// Initialize PostgreSQL repository for tests
	dbConfig := postgres.NewConfigFromEnv()
	repo, err := postgres.NewPostgresRepository(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize test database: %v", err)
	}
	database.Repo = repo
	db := repo

	// Get the singleton event broker
	eventBroker := events.GetBroker()

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
