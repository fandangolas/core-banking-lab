package database

import (
	"bank-api/internal/domain/models"
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

// Init initializes the repository with an in-memory implementation.
// Uses sync.Once to ensure it's only initialized once (singleton pattern).
func Init() {
	initOnce.Do(func() {
		Repo = NewInMemory()
	})
}
