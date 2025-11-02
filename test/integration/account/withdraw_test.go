package account

import (
	"bank-api/test/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdraw(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	router := testenv.SetupRouter()

	accountID := testenv.CreateAccount(t, router, "Nícolas")
	testenv.SetBalance(t, accountID, 5000)

	body := map[string]int{"amount": 3000}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(accountID)+"/withdraw", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	assert.Equal(t, float64(accountID), result["id"])
	assert.Equal(t, float64(2000), result["balance"])

	balance := testenv.GetBalance(t, router, accountID)
	assert.Equal(t, 2000, balance)
}

func TestWithdrawInvalidAmount(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	router := testenv.SetupRouter()

	accountID := testenv.CreateAccount(t, router, "Nícolas")
	testenv.SetBalance(t, accountID, 500)

	body := map[string]int{"amount": -100}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(accountID)+"/withdraw", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	testenv.AssertHasError(t, result)
}

func TestWithdrawInsufficientBalance(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	router := testenv.SetupRouter()

	accountID := testenv.CreateAccount(t, router, "Nícolas")
	testenv.SetBalance(t, accountID, 100)

	body := map[string]int{"amount": 500}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(accountID)+"/withdraw", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	testenv.AssertHasError(t, result)

	// Verify balance unchanged in database after failed withdrawal
	balance := testenv.GetBalance(t, router, accountID)
	assert.Equal(t, 100, balance, "Balance should remain unchanged after failed withdrawal")
}

func TestWithdrawNonexistentAccount(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	router := testenv.SetupRouter()

	body := map[string]int{"amount": 100}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/999/withdraw", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusNotFound, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	testenv.AssertHasError(t, result)
}

func TestConcurrentWithdraw(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	router := testenv.SetupRouter()

	accountID := testenv.CreateAccount(t, router, "ConcurrentWithdraw")
	testenv.SetBalance(t, accountID, 10000) // R$ 100,00

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

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Errorf("Erro no saque: %d", resp.Code)
			}
		}()
	}

	wg.Wait()

	finalBalance := testenv.GetBalance(t, router, accountID)
	expected := 0

	if finalBalance != expected {
		t.Errorf("ERRO DE CONCORRÊNCIA DETECTADO: saldo final %d, esperado %d", finalBalance, expected)
	} else {
		t.Logf("Saldo final correto: %d", finalBalance)
	}
}
