package database

import (
	"bank-api/internal/domain/models"
	"bank-api/internal/infrastructure/database/postgres"
	"log"
	"sync"
)

// Repository defines the required methods for persisting accounts.
type Repository interface {
	CreateAccount(owner string) int
	GetAccount(id int) (*models.Account, bool)
	UpdateAccount(acc *models.Account)
	Reset()
}

var (
	Repo     Repository
	initOnce sync.Once
)

// Init initializes the PostgreSQL repository.
// Uses sync.Once to ensure it's only initialized once (singleton pattern).
func Init() {
	initOnce.Do(func() {
		log.Println("Initializing PostgreSQL repository...")
		config := postgres.NewConfigFromEnv()
		repo, err := postgres.NewPostgresRepository(config.ConnectionString())
		if err != nil {
			log.Fatalf("Failed to initialize PostgreSQL repository: %v", err)
		}
		Repo = repo
		log.Println("PostgreSQL repository initialized successfully")
	})
}

// InitWithConnectionString directly initializes PostgreSQL repository with a connection string (for testing)
func InitWithConnectionString(connString string) error {
	repo, err := postgres.NewPostgresRepository(connString)
	if err != nil {
		return err
	}
	Repo = repo
	return nil
}
