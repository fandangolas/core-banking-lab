package account

import (
	"bank-api/src/db"
	"bank-api/tests/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestSimpleDeposit(t *testing.T) {
	testenv.SetupRouter()
	defer db.InMemory.Reset()

	accountID := testenv.CreateAccount(t, "Nicolas")

	body := map[string]int{"amount": 2500}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(accountID)+"/deposit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	testenv.SetupRouter().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("erro no depósito: %d", resp.Code)
	}

	balance := testenv.GetBalance(t, accountID)
	if balance != 2500 {
		t.Fatalf("esperado 2500, obtido %d", balance)
	}
}
