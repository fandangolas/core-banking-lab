package account

import (
	"bank-api/src/db"
	"bank-api/tests/integration/testenv"
	"testing"
)

func TestGetBalance(t *testing.T) {
	testenv.SetupRouter()
	defer db.InMemory.Reset()

	accountID := testenv.CreateAccount(t, "Nico")
	testenv.Deposit(t, accountID, 7500)

	balance := testenv.GetBalance(t, accountID)
	expected := 7500

	if balance != expected {
		t.Fatalf("Esperado saldo %d, obtido %d", expected, balance)
	}
}
