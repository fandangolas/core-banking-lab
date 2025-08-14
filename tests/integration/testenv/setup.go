package testenv

import (
	"bank-api/src/config"
	"bank-api/src/diplomat/database"
	"bank-api/src/diplomat/middleware"
	"bank-api/src/diplomat/routes"

	"github.com/gin-gonic/gin"
)

// SetupTestRouter creates a new router for testing with all routes and middleware
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	
	// Initialize database if not already done
	database.Init()
	
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
