package account

import (
	"bank-api/test/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	router := testenv.SetupRouter()

	body := map[string]string{"owner": "Alice"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusCreated, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	assert.Equal(t, "Alice", result["owner"])
	assert.NotZero(t, result["id"])

	// Verify account was persisted in database by retrieving balance via GET
	accountID := int(result["id"].(float64))
	balance := testenv.GetBalance(t, router, accountID)
	assert.Equal(t, 0, balance, "New account should have zero balance")
}

func TestCreateAccountInvalid(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	router := testenv.SetupRouter()

	body := map[string]string{"owner": ""}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
	testenv.AssertHasError(t, result)
}
