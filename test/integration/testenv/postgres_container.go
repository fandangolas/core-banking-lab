package testenv

import (
	"bank-api/internal/infrastructure/database"
	dbpostgres "bank-api/internal/infrastructure/database/postgres"
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testContainer     *postgres.PostgresContainer
	testContainerOnce sync.Once
	testContainerErr  error
)

// PostgresContainerConfig holds configuration for the test container
type PostgresContainerConfig struct {
	Database string
	Username string
	Password string
	Image    string
}

// DefaultPostgresConfig returns the default configuration for test containers
func DefaultPostgresConfig() PostgresContainerConfig {
	return PostgresContainerConfig{
		Database: "banking",
		Username: "banking",
		Password: "banking_secure_pass_2024",
		Image:    "postgres:16-alpine",
	}
}

// SetupPostgresContainer creates and starts a PostgreSQL testcontainer
// The container is automatically cleaned up when the test finishes
func SetupPostgresContainer(t *testing.T) (*postgres.PostgresContainer, string) {
	ctx := context.Background()
	cfg := DefaultPostgresConfig()

	// Create PostgreSQL container with configuration
	container, err := postgres.Run(ctx,
		cfg.Image,
		postgres.WithDatabase(cfg.Database),
		postgres.WithUsername(cfg.Username),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err, "Failed to start PostgreSQL testcontainer")

	// Automatic cleanup when test completes
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL testcontainer: %v", err)
		}
	})

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "Failed to get connection string from testcontainer")

	return container, connStr
}

// SetupPostgresContainerWithEnv creates a PostgreSQL testcontainer and sets environment variables
// This is useful for code that reads configuration from environment variables
// Deprecated: Use SetupIntegrationTest instead for better performance and proper schema initialization
func SetupPostgresContainerWithEnv(t *testing.T) *postgres.PostgresContainer {
	ctx := context.Background()
	cfg := DefaultPostgresConfig()

	// Create container with init script
	container, err := postgres.Run(ctx,
		cfg.Image,
		postgres.WithDatabase(cfg.Database),
		postgres.WithUsername(cfg.Username),
		postgres.WithPassword(cfg.Password),
		postgres.WithInitScripts("../../../internal/infrastructure/database/postgres/migrations/000001_init_schema.up.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err, "Failed to start PostgreSQL testcontainer")

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL testcontainer: %v", err)
		}
	})

	// Parse connection details and set environment variables
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Set environment variables for code that reads from env
	t.Setenv("DB_HOST", host)
	t.Setenv("DB_PORT", port.Port())
	t.Setenv("DB_NAME", cfg.Database)
	t.Setenv("DB_USER", cfg.Username)
	t.Setenv("DB_PASSWORD", cfg.Password)
	t.Setenv("DB_SSLMODE", "disable")

	connStr, _ := container.ConnectionString(ctx, "sslmode=disable")
	t.Logf("PostgreSQL testcontainer ready: %s", connStr)

	return container
}

// SetupIntegrationTest initializes the PostgreSQL testcontainer once and sets up the database repository
// This function should be called at the beginning of each integration test
// The container is shared across all tests and cleaned up automatically
func SetupIntegrationTest(t *testing.T) {
	// Initialize container once
	testContainerOnce.Do(func() {
		ctx := context.Background()
		cfg := DefaultPostgresConfig()

		// Create PostgreSQL container with init script
		container, err := postgres.Run(ctx,
			cfg.Image,
			postgres.WithDatabase(cfg.Database),
			postgres.WithUsername(cfg.Username),
			postgres.WithPassword(cfg.Password),
			postgres.WithInitScripts("../../../internal/infrastructure/database/postgres/migrations/000001_init_schema.up.sql"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*time.Second),
			),
		)
		if err != nil {
			testContainerErr = fmt.Errorf("failed to start PostgreSQL testcontainer: %w", err)
			return
		}

		testContainer = container

		// Get connection details
		host, err := container.Host(ctx)
		if err != nil {
			testContainerErr = fmt.Errorf("failed to get container host: %w", err)
			return
		}

		port, err := container.MappedPort(ctx, "5432")
		if err != nil {
			testContainerErr = fmt.Errorf("failed to get container port: %w", err)
			return
		}

		// Create config for database
		dbConfig := &dbpostgres.Config{
			Host:              host,
			Port:              5432,
			Database:          cfg.Database,
			User:              cfg.Username,
			Password:          cfg.Password,
			SSLMode:           "disable",
			MaxOpenConns:      25,
			MaxIdleConns:      5,
			ConnMaxLifetime:   "30m",
			ConnMaxIdleTime:   "5m",
			HealthCheckPeriod: "1m",
		}

		// Override port with actual mapped port
		dbConfig.Port = port.Int()

		// Initialize repository
		repo, err := dbpostgres.NewPostgresRepository(dbConfig)
		if err != nil {
			testContainerErr = fmt.Errorf("failed to create repository: %w", err)
			return
		}

		database.Repo = repo

		connStr, _ := container.ConnectionString(ctx, "sslmode=disable")
		log.Printf("PostgreSQL testcontainer initialized: %s", connStr)
	})

	// Check if initialization failed
	require.NoError(t, testContainerErr, "Failed to initialize test container")

	// Reset database before each test
	if database.Repo != nil {
		database.Repo.Reset()
	}
}
