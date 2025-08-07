package database

import "bank-api/src/models"

// Repository defines the required methods for persisting accounts.
type Repository interface {
	CreateAccount(owner string) int
	GetAccount(id int) (*models.Account, bool)
	UpdateAccount(acc *models.Account)
	Reset()
}

var Repo Repository

// Init initializes the repository with an in-memory implementation.
func Init() {
	Repo = NewInMemory()
}
