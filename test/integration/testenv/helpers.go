package testenv

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"bank-api/internal/domain/account"
	"bank-api/internal/infrastructure/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func CreateAccount(t *testing.T, r *gin.Engine, owner string) int {
	body := map[string]interface{}{"owner": owner}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("erro ao criar conta: %d", resp.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	return int(result["id"].(float64))
}

func GetBalance(t *testing.T, r *gin.Engine, id int) int {
	req := httptest.NewRequest("GET", "/accounts/"+strconv.Itoa(id)+"/balance", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("erro ao consultar saldo: %d", resp.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	return int(result["balance"].(float64))
}

func Deposit(t *testing.T, r *gin.Engine, id int, amount int) string {
	body := map[string]int{"amount": amount}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(id)+"/deposit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	// Now expects 202 Accepted for async processing
	if resp.Code != http.StatusAccepted {
		t.Fatalf("erro no depósito: %d", resp.Code)
	}

	// Return operation ID for tracking
	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	if opID, ok := result["operation_id"].(string); ok {
		return opID
	}
	return ""
}

func Withdraw(t *testing.T, r *gin.Engine, id int, amount int) {
	body := map[string]int{"amount": amount}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(id)+"/withdraw", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("erro no saque: %d", resp.Code)
	}
}

// AssertHasError checks if the response has an error message in either the new format (message) or old format (error)
func AssertHasError(t *testing.T, result map[string]interface{}) {
	if message, ok := result["message"]; ok {
		assert.NotEmpty(t, message, "Expected error message to be present")
	} else if errorMsg, ok := result["error"]; ok {
		assert.NotEmpty(t, errorMsg, "Expected error message to be present")
	} else {
		t.Error("No error message found in response")
	}
}

// SetBalance directly sets an account balance for test setup purposes
// This bypasses the async deposit mechanism and is only for test fixtures
func SetBalance(t *testing.T, accountID int, amount int) {
	acc, ok := database.Repo.GetAccount(accountID)
	if !ok {
		t.Fatalf("account not found: %d", accountID)
	}

	// Add the amount to the account
	if err := domain.AddAmount(acc, amount); err != nil {
		t.Fatalf("failed to add amount: %v", err)
	}

	database.Repo.UpdateAccount(acc)
}
