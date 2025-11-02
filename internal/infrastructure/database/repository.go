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

	// Atomic operations for concurrency safety
	AtomicWithdraw(accountID int, amount int) (*models.Account, error)
	AtomicTransfer(fromID int, toID int, amount int) (*models.Account, *models.Account, error)
}

var (
	// Repo is the global repository instance, initialized by the components layer
	Repo Repository
)

// InitWithConfig directly initializes PostgreSQL repository with a config (for testing)
func InitWithConfig(cfg *postgres.Config) error {
	repo, err := postgres.NewPostgresRepository(cfg)
	if err != nil {
		return err
	}
	Repo = repo
	return nil
}
