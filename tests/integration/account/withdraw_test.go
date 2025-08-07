package account

import (
	"bank-api/src/db"
	"bank-api/tests/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
)

func TestWithdraw(t *testing.T) {
	testenv.SetupRouter()
	defer db.InMemory.Reset()

	accountID := testenv.CreateAccount(t, "Nícolas")
	testenv.Deposit(t, accountID, 5000)

	body := map[string]int{"amount": 3000}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(accountID)+"/withdraw", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	testenv.SetupRouter().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Erro no saque: %d", resp.Code)
	}

	balance := testenv.GetBalance(t, accountID)
	expected := 2000
	if balance != expected {
		t.Fatalf("Esperado saldo %d, obtido %d", expected, balance)
	}
}

func TestConcurrentWithdraw(t *testing.T) {
	testenv.SetupRouter()
	defer db.InMemory.Reset()

	accountID := testenv.CreateAccount(t, "ConcurrentWithdraw")
	testenv.Deposit(t, accountID, 10000) // R$ 100,00

	var wg sync.WaitGroup
	n := 100
	amount := 100 // R$ 1,00 por saque
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			body := map[string]int{"amount": amount}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(accountID)+"/withdraw", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			testenv.SetupRouter().ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Errorf("Erro no saque: %d", resp.Code)
			}
		}()
	}

	wg.Wait()

	finalBalance := testenv.GetBalance(t, accountID)
	expected := 0

	if finalBalance != expected {
		t.Errorf("ERRO DE CONCORRÊNCIA DETECTADO: saldo final %d, esperado %d", finalBalance, expected)
	} else {
		t.Logf("Saldo final correto: %d", finalBalance)
	}
}
