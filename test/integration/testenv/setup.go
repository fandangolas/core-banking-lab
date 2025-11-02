package testenv

import (
	"bank-api/internal/api/middleware"
	"bank-api/internal/api/routes"
	"bank-api/internal/config"
	"bank-api/internal/infrastructure/database"
	"bank-api/internal/infrastructure/messaging"

	"github.com/gin-gonic/gin"
)

// handlerContainer is a simple implementation of handlers.HandlerDependencies for tests
type handlerContainer struct {
	db        database.Repository
	publisher messaging.EventPublisher
}

func (h *handlerContainer) GetDatabase() database.Repository {
	return h.db
}

func (h *handlerContainer) GetEventPublisher() messaging.EventPublisher {
	return h.publisher
}

// SetupTestRouter creates a new router for testing with all routes and middleware
// Note: Database initialization is now handled per-test using testcontainers
func SetupTestRouter() *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router for each test
	router := gin.Default()

	// Create minimal config for CORS
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"*"},
		},
	}

	// Apply middleware
	router.Use(middleware.CORS(cfg))

	// Create test container with no-op event publisher
	container := &handlerContainer{
		db:        database.Repo,
		publisher: messaging.NewNoOpEventPublisher(),
	}

	// Register routes with container
	routes.RegisterRoutes(router, container)

	return router
}

// SetupTestRouterWithEventPublisher creates a router with event publisher
func SetupTestRouterWithEventPublisher(publisher messaging.EventPublisher) *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router for each test
	router := gin.Default()

	// Create minimal config for CORS
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"*"},
		},
	}

	// Apply middleware
	router.Use(middleware.CORS(cfg))

	// Create test container with provided event publisher
	container := &handlerContainer{
		db:        database.Repo,
		publisher: publisher,
	}

	// Register routes with container
	routes.RegisterRoutes(router, container)

	return router
}

// SetupRouter is maintained for backward compatibility
func SetupRouter() *gin.Engine {
	return SetupTestRouter()
}
