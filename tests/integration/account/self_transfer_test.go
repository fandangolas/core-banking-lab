package account

import (
	"bank-api/src/diplomat/database"
	"bank-api/tests/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransferToSameAccount(t *testing.T) {
	router := testenv.SetupRouter()
	defer database.Repo.Reset()

	accountID := testenv.CreateAccount(t, router, "Self")
	testenv.Deposit(t, router, accountID, 1000)

	body := map[string]int{
		"from":   accountID,
		"to":     accountID,
		"amount": 500,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	assert.NotEmpty(t, result["error"])

	balance := testenv.GetBalance(t, router, accountID)
	assert.Equal(t, 1000, balance)
}
