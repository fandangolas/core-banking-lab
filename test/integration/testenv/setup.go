package testenv

import (
	"bank-api/internal/api/middleware"
	"bank-api/internal/api/routes"
	"bank-api/internal/config"
	"bank-api/internal/infrastructure/database"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	setupOnce sync.Once
)

// SetupTestRouter creates a new router for testing with all routes and middleware
func SetupTestRouter() *gin.Engine {
	// Ensure database is initialized only once across all tests
	setupOnce.Do(func() {
		gin.SetMode(gin.TestMode)
		database.Init()
	})

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

	// Register routes
	routes.RegisterRoutes(router)

	return router
}

// SetupRouter is maintained for backward compatibility
func SetupRouter() *gin.Engine {
	return SetupTestRouter()
}
