package components

import (
	"bank-api/src/config"
	"bank-api/src/diplomat/database"
	"bank-api/src/diplomat/events"
	"bank-api/src/diplomat/middleware"
	"bank-api/src/diplomat/routes"
	"bank-api/src/logging"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Container holds all application components and their dependencies
type Container struct {
	Config      *config.Config
	Logger      *logging.Logger
	Database    database.Repository
	EventBroker *events.Broker
	Router      *gin.Engine
	Server      *http.Server
}

var (
	instance     *Container
	instanceOnce sync.Once
	instanceErr  error
)

// GetInstance returns the singleton container instance.
// Uses sync.Once to ensure it's only initialized once.
func GetInstance() (*Container, error) {
	instanceOnce.Do(func() {
		instance, instanceErr = newContainer()
	})
	return instance, instanceErr
}

// New creates and initializes all application components.
// For backward compatibility, this calls GetInstance.
func New() (*Container, error) {
	return GetInstance()
}

// newContainer creates a new container instance (internal use only)
func newContainer() (*Container, error) {
	container := &Container{}

	// Initialize configuration
	if err := container.initConfig(); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	// Initialize logger
	if err := container.initLogger(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize database
	if err := container.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize event broker
	if err := container.initEventBroker(); err != nil {
		return nil, fmt.Errorf("failed to initialize event broker: %w", err)
	}

	// Initialize router and server
	if err := container.initServer(); err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	logging.Info("All components initialized successfully", nil)
	return container, nil
}

// initConfig loads the application configuration
func (c *Container) initConfig() error {
	c.Config = config.Load()
	return nil
}

// initLogger sets up the logging system
func (c *Container) initLogger() error {
	logging.Init(c.Config)
	c.Logger = &logging.Logger{}

	logging.Info("Logger initialized", map[string]interface{}{
		"level": c.Config.Logging.Level,
	})
	return nil
}

// initDatabase sets up the database connection
func (c *Container) initDatabase() error {
	database.Init()
	c.Database = database.Repo

	logging.Info("Database initialized", map[string]interface{}{
		"type": "in-memory",
	})
	return nil
}

// initEventBroker sets up the event broadcasting system
func (c *Container) initEventBroker() error {
	// Get the singleton event broker instance
	c.EventBroker = events.GetBroker()

	logging.Info("Event broker initialized", nil)
	return nil
}

// initServer sets up the HTTP server with all middleware and routes
func (c *Container) initServer() error {
	// Setup Gin router
	// Set Gin mode based on environment
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	c.Router = gin.Default()

	// Apply global middleware
	c.Router.Use(middleware.CORS(c.Config))

	// Register all routes
	routes.RegisterRoutes(c.Router)

	// Create HTTP server
	c.Server = &http.Server{
		Addr:           ":" + c.Config.Server.Port,
		Handler:        c.Router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	logging.Info("HTTP server configured", map[string]interface{}{
		"port": c.Config.Server.Port,
	})
	return nil
}

// Start begins serving HTTP requests
func (c *Container) Start() error {
	logging.Info("Starting HTTP server", map[string]interface{}{
		"address": c.Server.Addr,
	})

	// Start server in a goroutine
	go func() {
		if err := c.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("Server failed to start", err, nil)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	c.waitForShutdown()
	return nil
}

// waitForShutdown handles graceful shutdown
func (c *Container) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Info("Shutting down server...", nil)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.Shutdown(ctx); err != nil {
		logging.Error("Server forced to shutdown", err, nil)
	}

	logging.Info("Server shutdown complete", nil)
}

// Shutdown gracefully stops all components
func (c *Container) Shutdown(ctx context.Context) error {
	// Shutdown HTTP server
	if err := c.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// Here we could add cleanup for other components if needed
	// For example: close database connections, flush metrics, etc.

	return nil
}

// GetDatabase returns the database repository
func (c *Container) GetDatabase() database.Repository {
	return c.Database
}

// GetEventBroker returns the event broker
func (c *Container) GetEventBroker() *events.Broker {
	return c.EventBroker
}

// GetConfig returns the configuration
func (c *Container) GetConfig() *config.Config {
	return c.Config
}

// GetRouter returns the Gin router
func (c *Container) GetRouter() *gin.Engine {
	return c.Router
}
