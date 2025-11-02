package testenv

import (
	"bank-api/internal/api/middleware"
	"bank-api/internal/api/routes"
	"bank-api/internal/config"

	"github.com/gin-gonic/gin"
)

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

	// Register routes
	routes.RegisterRoutes(router)

	return router
}

// SetupRouter is maintained for backward compatibility
func SetupRouter() *gin.Engine {
	return SetupTestRouter()
}
