package db

import (
	"bank-api/src/models"
	"sync"
)

type InMemoryDB struct {
	accounts map[int]*models.Account
	nextID   int

	mu sync.RWMutex
}

var InMemory *InMemoryDB

func Init() {
	InMemory = &InMemoryDB{
		accounts: make(map[int]*models.Account),
		nextID:   1,
	}
}

func (db *InMemoryDB) CreateAccount(owner string) int {
	db.mu.Lock()
	defer db.mu.Unlock()

	id := db.nextID
	db.nextID++

	db.accounts[id] = &models.Account{
		Id:      id,
		Owner:   owner,
		Balance: 0.0,
	}

	return id
}

func (db *InMemoryDB) GetAccount(id int) (*models.Account, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	account, ok := db.accounts[id]
	return account, ok
}

func (db *InMemoryDB) UpdateAccount(acc *models.Account) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.accounts[acc.Id] = acc
}

func (db *InMemoryDB) Reset() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.accounts = make(map[int]*models.Account)
	db.nextID = 1
}
