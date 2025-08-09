package account

import (
	"bank-api/src/diplomat/database"
	"bank-api/tests/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleDeposit(t *testing.T) {
	router := testenv.SetupRouter()
	defer database.Repo.Reset()

	accountID := testenv.CreateAccount(t, router, "Nicolas")

	body := map[string]int{"amount": 2500}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(accountID)+"/deposit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	assert.Equal(t, float64(accountID), result["id"])
	assert.Equal(t, float64(2500), result["balance"])

	balance := testenv.GetBalance(t, router, accountID)
	assert.Equal(t, 2500, balance)
}

func TestDepositInvalidAmount(t *testing.T) {
	router := testenv.SetupRouter()
	defer database.Repo.Reset()

	accountID := testenv.CreateAccount(t, router, "Nicolas")

	body := map[string]int{"amount": -100}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(accountID)+"/deposit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	assert.NotEmpty(t, result["error"])
}

func TestDepositNonexistentAccount(t *testing.T) {
	router := testenv.SetupRouter()
	defer database.Repo.Reset()

	body := map[string]int{"amount": 100}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/999/deposit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusNotFound, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	assert.NotEmpty(t, result["error"])
}
