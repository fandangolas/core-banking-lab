package database

import (
	"bank-api/src/models"
	"sync"
)

type InMemory struct {
	accounts map[int]*models.Account
	nextID   int
	mu       sync.RWMutex
}

// NewInMemory creates a new in-memory repository instance.
func NewInMemory() Repository {
	return &InMemory{
		accounts: make(map[int]*models.Account),
		nextID:   1,
	}
}

func (db *InMemory) CreateAccount(owner string) int {
	db.mu.Lock()
	defer db.mu.Unlock()

	id := db.nextID
	db.nextID++

	db.accounts[id] = &models.Account{
		Id:      id,
		Owner:   owner,
		Balance: 0,
	}

	return id
}

func (db *InMemory) GetAccount(id int) (*models.Account, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	account, ok := db.accounts[id]
	return account, ok
}

func (db *InMemory) UpdateAccount(acc *models.Account) {
	// Note: This method is actually not needed since we work with
	// pointers to the accounts directly. The mutex locks in domain
	// layer already protect concurrent access to account fields.
	// This is kept for interface compliance but doesn't need to do anything
	// as we're modifying the same account reference that's in the map.
}

func (db *InMemory) Reset() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.accounts = make(map[int]*models.Account)
	db.nextID = 1
}

// GetAccountCount returns the number of accounts for metrics
func (db *InMemory) GetAccountCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.accounts)
}
