package account

import (
	"bank-api/internal/infrastructure/database"
	"bank-api/test/integration/testenv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBalance(t *testing.T) {
	router := testenv.SetupRouter()
	defer database.Repo.Reset()

	accountID := testenv.CreateAccount(t, router, "Nico")
	testenv.Deposit(t, router, accountID, 7500)

	balance := testenv.GetBalance(t, router, accountID)
	assert.Equal(t, 7500, balance)
}

func TestGetBalanceNonexistentAccount(t *testing.T) {
	router := testenv.SetupRouter()
	defer database.Repo.Reset()

	req := httptest.NewRequest("GET", "/accounts/999/balance", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusNotFound, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	testenv.AssertHasError(t, result)
}
