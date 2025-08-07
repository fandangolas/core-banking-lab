package domain

import (
	"bank-api/src/models"
	"errors"
)

func withAccountLock(acc *models.Account, fn func()) {
	acc.Mu.Lock()
	defer acc.Mu.Unlock()
	fn()
}

func AddAmount(acc *models.Account, amount int) error {
	if amount <= 0 {
		return errors.New("invalid amount")
	}

	withAccountLock(acc, func() {
		acc.Balance += amount
	})

	return nil
}

func RemoveAmount(acc *models.Account, amount int) error {
	if amount <= 0 {
		return errors.New("invalid amount")
	}

	var err error
	withAccountLock(acc, func() {
		if acc.Balance-amount < 0 {
			err = errors.New("invalid amount, greater than balance")
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
