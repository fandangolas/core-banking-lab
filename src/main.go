package main

// Core Banking API with comprehensive monitoring stack
import (
	"bank-api/src/components"
	"bank-api/src/logging"
	"log"
)

func main() {
	// Initialize all application components
	container, err := components.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	logging.Info("Bank API initialized successfully", map[string]interface{}{
		"version":     "1.0.0",
		"environment": container.GetConfig().Environment,
		"port":        container.GetConfig().Server.Port,
	})

	// Start the server (this will block until shutdown)
	if err := container.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
