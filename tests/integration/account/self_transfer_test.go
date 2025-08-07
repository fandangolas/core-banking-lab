package account

import (
	"bank-api/src/db"
	"bank-api/tests/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTransferToSameAccount(t *testing.T) {
	testenv.SetupRouter()
	defer db.InMemory.Reset()

	accountID := testenv.CreateAccount(t, "Self")
	testenv.Deposit(t, accountID, 1000)

	body := map[string]int{
		"from":   accountID,
		"to":     accountID,
		"amount": 500,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	testenv.SetupRouter().ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("esperado status %d, obtido %d", http.StatusBadRequest, resp.Code)
	}

	balance := testenv.GetBalance(t, accountID)
	expected := 1000
	if balance != expected {
		t.Fatalf("esperado saldo %d, obtido %d", expected, balance)
	}
}
