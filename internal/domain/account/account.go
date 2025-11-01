package domain

import (
	"bank-api/internal/domain/models"
	"bank-api/internal/pkg/validation"
	"errors"
)

func withAccountLock(acc *models.Account, fn func()) {
	acc.Mu.Lock()
	defer acc.Mu.Unlock()
	fn()
}

func AddAmount(acc *models.Account, amount int) error {
	if err := validation.ValidateAmount(amount); err != nil {
		return err
	}

	withAccountLock(acc, func() {
		acc.Balance += amount
	})

	return nil
}

func RemoveAmount(acc *models.Account, amount int) error {
	if err := validation.ValidateAmount(amount); err != nil {
		return err
	}

	var err error
	withAccountLock(acc, func() {
		if acc.Balance-amount < 0 {
			err = errors.New("insufficient balance")
			return
		}

		acc.Balance -= amount
	})

	return err
}

func GetBalance(acc *models.Account) int {
	var balance int
	withAccountLock(acc, func() {
		balance = acc.Balance
	})
	return balance
}
