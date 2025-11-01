package database

import (
	"bank-api/internal/domain/models"
	"bank-api/internal/infrastructure/database/postgres"
)

// Repository defines the required methods for persisting accounts.
type Repository interface {
	CreateAccount(owner string) int
	GetAccount(id int) (*models.Account, bool)
	UpdateAccount(acc *models.Account)
	Reset()
}

var (
	// Repo is the global repository instance, initialized by the components layer
	Repo Repository
)

// InitWithConnectionString directly initializes PostgreSQL repository with a connection string (for testing)
func InitWithConnectionString(connString string) error {
	repo, err := postgres.NewPostgresRepository(connString)
	if err != nil {
		return err
	}
	Repo = repo
	return nil
}
