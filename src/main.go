package main

import (
	"bank-api/src/config"
	"bank-api/src/diplomat/database"
	"bank-api/src/diplomat/middleware"
	"bank-api/src/diplomat/routes"
	"bank-api/src/logging"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logging
	logging.Init(cfg)
	logging.Info("Starting bank-api server", map[string]interface{}{
		"port": cfg.Server.Port,
		"host": cfg.Server.Host,
	})

	// Initialize database
	database.Init()

	// Setup router with middleware - fixed CORS with relative URLs
	router := gin.Default()
	router.Use(middleware.CORS(cfg))
	//router.Use(middleware.RateLimit(cfg))

	// Register routes
	routes.RegisterRoutes(router)

	// Start server
	addr := ":" + cfg.Server.Port
	logging.Info("Server listening", map[string]interface{}{"address": addr})
	router.Run(addr)
}
