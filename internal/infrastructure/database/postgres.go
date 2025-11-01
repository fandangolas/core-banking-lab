package database

import "bank-api/internal/domain/models"

// Postgres is a placeholder for a PostgreSQL-backed repository.
type Postgres struct{}

// NewPostgres creates a new PostgreSQL repository instance.
func NewPostgres() Repository {
	return &Postgres{}
}

func (pg *Postgres) CreateAccount(owner string) int {
	// TODO: implement PostgreSQL storage
	return 0
}

func (pg *Postgres) GetAccount(id int) (*models.Account, bool) {
	return nil, false
}

func (pg *Postgres) UpdateAccount(acc *models.Account) {}

func (pg *Postgres) Reset() {}
