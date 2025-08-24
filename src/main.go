package main

import (
	"bank-api/src/components"
	"bank-api/src/logging"
	"log"
	"os"
	"runtime"
	"runtime/debug"
)

func main() {
	// Performance optimizations
	optimizeRuntime()
	
	container, err := components.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	logging.Info("Bank API initialized successfully", map[string]interface{}{
		"version":     "1.0.0",
		"environment": container.GetConfig().Environment,
		"port":        container.GetConfig().Server.Port,
		"gomaxprocs":  runtime.GOMAXPROCS(0),
		"gogc":        os.Getenv("GOGC"),
	})

	if err := container.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func optimizeRuntime() {
	// Use all available CPU cores
	if os.Getenv("GOMAXPROCS") == "" {
		numCPU := runtime.NumCPU()
		runtime.GOMAXPROCS(numCPU)
		log.Printf("Setting GOMAXPROCS to %d", numCPU)
	}
	
	// Reduce GC frequency for better throughput
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(400) // Run GC less frequently
		log.Printf("Setting GOGC to 400")
	}
	
	// Pre-allocate memory to reduce GC pressure
	ballast := make([]byte, 100<<20) // 100MB ballast
	runtime.KeepAlive(ballast)
}
