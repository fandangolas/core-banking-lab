package main

import (
	"bank-api/internal/pkg/components"
	"bank-api/internal/pkg/logging"
	"log"
)

func main() {
	container, err := components.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	logging.Info("Bank API initialized successfully", map[string]interface{}{
		"version":     "1.0.0",
		"environment": container.GetConfig().Environment,
		"port":        container.GetConfig().Server.Port,
	})

	if err := container.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
