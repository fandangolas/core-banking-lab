package domain_test

import (
	"bank-api/src/domain"
	"bank-api/src/models"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAccount(balance int) *models.Account {
	return &models.Account{
		Id:      1,
		Owner:   "Test",
		Balance: balance,
		Mu:      sync.Mutex{},
	}
}

func TestAddAmount(t *testing.T) {
	tests := []struct {
		name    string
		initial int
		amount  int
		want    int
		wantErr bool
	}{
		{"valid", 1000, 500, 1500, false},
		{"invalid", 1000, -100, 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := newTestAccount(tt.initial)
			err := domain.AddAmount(acc, tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, acc.Balance)
		})
	}
}

func TestRemoveAmount(t *testing.T) {
	tests := []struct {
		name    string
		initial int
		amount  int
		want    int
		wantErr bool
	}{
		{"valid", 1000, 300, 700, false},
		{"insufficient", 200, 500, 200, true},
		{"invalid", 200, -50, 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := newTestAccount(tt.initial)
			err := domain.RemoveAmount(acc, tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, acc.Balance)
		})
	}
}

func TestGetBalance(t *testing.T) {
	acc := newTestAccount(500)
	assert.Equal(t, 500, domain.GetBalance(acc))
}

func TestConcurrentAddAmount(t *testing.T) {
	acc := newTestAccount(0)
	var wg sync.WaitGroup
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			err := domain.AddAmount(acc, 1)
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.Equal(t, n, domain.GetBalance(acc))
}

func TestConcurrentRemoveAmount(t *testing.T) {
	acc := newTestAccount(500)
	var wg sync.WaitGroup
	n := 100
	amount := 2
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			err := domain.RemoveAmount(acc, amount)
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.Equal(t, 500-n*amount, domain.GetBalance(acc))
}
