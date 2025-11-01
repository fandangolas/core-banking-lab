package account

import (
	"bank-api/internal/infrastructure/database"
	"bank-api/test/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransferSuccess(t *testing.T) {
	router := testenv.SetupRouter()
	defer database.Repo.Reset()

	from := testenv.CreateAccount(t, router, "From")
	to := testenv.CreateAccount(t, router, "To")
	testenv.Deposit(t, router, from, 1000)

	body := map[string]int{"from": from, "to": to, "amount": 300}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	assert.Equal(t, float64(from), result["from_id"])
	assert.Equal(t, float64(to), result["to_id"])
	assert.Equal(t, float64(700), result["from_balance"])
	assert.Equal(t, float64(300), result["to_balance"])

	assert.Equal(t, 700, testenv.GetBalance(t, router, from))
	assert.Equal(t, 300, testenv.GetBalance(t, router, to))
}

func TestTransferNonexistentAccount(t *testing.T) {
	router := testenv.SetupRouter()
	defer database.Repo.Reset()

	from := testenv.CreateAccount(t, router, "From")
	testenv.Deposit(t, router, from, 100)

	body := map[string]int{"from": from, "to": 999, "amount": 50}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusNotFound, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	testenv.AssertHasError(t, result)

	// Verify source account balance unchanged in database after failed transfer
	balance := testenv.GetBalance(t, router, from)
	assert.Equal(t, 100, balance, "Source account balance should remain unchanged after failed transfer")
}
