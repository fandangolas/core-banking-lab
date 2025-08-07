package logic_test

import (
	"bank-api/src/logic"
	"bank-api/src/models"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestAccount(balance int) *models.Account {
	return &models.Account{
		Id:      1,
		Owner:   "Test",
		Balance: balance,
		Mu:      sync.Mutex{},
	}
}

func TestAddAmount_Valid(t *testing.T) {
	account := newTestAccount(1000)

	err := logic.AddAmount(account, 500)

	assert.NoError(t, err)
	assert.Equal(t, 1500, account.Balance)
}

func TestAddAmount_Invalid(t *testing.T) {
	account := newTestAccount(1000)

	err := logic.AddAmount(account, -100)

	assert.Error(t, err)
	assert.Equal(t, 1000, account.Balance)
}

func TestRemoveAmount_Valid(t *testing.T) {
	account := newTestAccount(1000)

	err := logic.RemoveAmount(account, 300)

	assert.NoError(t, err)
	assert.Equal(t, 700, account.Balance)
}

func TestRemoveAmount_InsufficientBalance(t *testing.T) {
	account := newTestAccount(200)

	err := logic.RemoveAmount(account, 500)

	assert.Error(t, err)
	assert.Equal(t, 200, account.Balance)
}

func TestRemoveAmount_InvalidAmount(t *testing.T) {
	account := newTestAccount(200)

	err := logic.RemoveAmount(account, -50)

	assert.Error(t, err)
	assert.Equal(t, 200, account.Balance)
}
