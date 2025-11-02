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

// TestAccountCreatedEventPublished verifies that AccountCreatedEvent is published when creating an account
func TestAccountCreatedEventPublished(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()
	eventPublisher := container.GetEventPublisher()

	// Create account
	body := map[string]string{"owner": "Alice"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusCreated, resp.Code)

	// Verify event was captured
	events := eventPublisher.GetAccountCreatedEvents()
	require.Len(t, events, 1, "Expected exactly one AccountCreatedEvent")

	event := events[0]
	assert.Equal(t, "Alice", event.Owner)
	assert.NotZero(t, event.AccountID)
	assert.False(t, event.Timestamp.IsZero())
}

// TestDepositEventPublished verifies that DepositCompletedEvent is published
func TestDepositEventPublished(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()
	eventPublisher := container.GetEventPublisher()

	// Create account first
	accountID := testenv.CreateAccount(t, router, "Bob")

	// Make deposit
	body := map[string]int{"amount": 1000}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/1/deposit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	// Verify deposit event was captured
	events := eventPublisher.GetDepositCompletedEvents()
	require.Len(t, events, 1, "Expected exactly one DepositCompletedEvent")

	event := events[0]
	assert.Equal(t, accountID, event.AccountID)
	assert.Equal(t, 1000, event.Amount)
	assert.Equal(t, 1000, event.BalanceAfter)
	assert.False(t, event.Timestamp.IsZero())
}

// TestWithdrawalEventPublished verifies that WithdrawalCompletedEvent is published
func TestWithdrawalEventPublished(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()
	eventPublisher := container.GetEventPublisher()

	// Create account and deposit funds
	accountID := testenv.CreateAccount(t, router, "Charlie")
	testenv.Deposit(t, router, accountID, 2000)

	// Reset events to clear the deposit event
	eventPublisher.Reset()

	// Make withdrawal
	body := map[string]int{"amount": 500}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/1/withdraw", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	// Verify withdrawal event was captured
	events := eventPublisher.GetWithdrawalCompletedEvents()
	require.Len(t, events, 1, "Expected exactly one WithdrawalCompletedEvent")

	event := events[0]
	assert.Equal(t, accountID, event.AccountID)
	assert.Equal(t, 500, event.Amount)
	assert.Equal(t, 1500, event.BalanceAfter)
	assert.False(t, event.Timestamp.IsZero())
}

// TestTransferEventPublished verifies that TransferCompletedEvent is published
func TestTransferEventPublished(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()
	eventPublisher := container.GetEventPublisher()

	// Create two accounts
	fromID := testenv.CreateAccount(t, router, "David")
	toID := testenv.CreateAccount(t, router, "Eve")

	// Deposit funds into from account
	testenv.Deposit(t, router, fromID, 3000)

	// Reset events to clear previous events
	eventPublisher.Reset()

	// Make transfer
	body := map[string]int{
		"from":   fromID,
		"to":     toID,
		"amount": 1200,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	// Verify transfer event was captured
	events := eventPublisher.GetTransferCompletedEvents()
	require.Len(t, events, 1, "Expected exactly one TransferCompletedEvent")

	event := events[0]
	assert.Equal(t, fromID, event.FromAccountID)
	assert.Equal(t, toID, event.ToAccountID)
	assert.Equal(t, 1200, event.Amount)
	assert.Equal(t, 1800, event.FromBalanceAfter, "From account should have 3000 - 1200 = 1800")
	assert.Equal(t, 1200, event.ToBalanceAfter, "To account should have 0 + 1200 = 1200")
	assert.False(t, event.Timestamp.IsZero())
}

// TestMultipleOperationsEventSequence verifies correct event capture for a sequence of operations
func TestMultipleOperationsEventSequence(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()
	eventPublisher := container.GetEventPublisher()

	// Create account
	accountID := testenv.CreateAccount(t, router, "Frank")

	// Perform multiple operations
	testenv.Deposit(t, router, accountID, 1000)
	testenv.Deposit(t, router, accountID, 500)
	testenv.Withdraw(t, router, accountID, 300)

	// Verify all events were captured
	accountEvents := eventPublisher.GetAccountCreatedEvents()
	depositEvents := eventPublisher.GetDepositCompletedEvents()
	withdrawalEvents := eventPublisher.GetWithdrawalCompletedEvents()

	assert.Len(t, accountEvents, 1, "Expected 1 account creation event")
	assert.Len(t, depositEvents, 2, "Expected 2 deposit events")
	assert.Len(t, withdrawalEvents, 1, "Expected 1 withdrawal event")

	// Verify event order and balances
	assert.Equal(t, 1000, depositEvents[0].Amount)
	assert.Equal(t, 1000, depositEvents[0].BalanceAfter)

	assert.Equal(t, 500, depositEvents[1].Amount)
	assert.Equal(t, 1500, depositEvents[1].BalanceAfter)

	assert.Equal(t, 300, withdrawalEvents[0].Amount)
	assert.Equal(t, 1200, withdrawalEvents[0].BalanceAfter)
}

// TestEventCaptureReset verifies that Reset() clears all captured events
func TestEventCaptureReset(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()
	eventPublisher := container.GetEventPublisher()

	// Create account and make deposit
	accountID := testenv.CreateAccount(t, router, "Grace")
	testenv.Deposit(t, router, accountID, 1000)

	// Verify events were captured
	assert.Len(t, eventPublisher.GetAccountCreatedEvents(), 1)
	assert.Len(t, eventPublisher.GetDepositCompletedEvents(), 1)

	// Reset event capture
	eventPublisher.Reset()

	// Verify all events were cleared
	assert.Len(t, eventPublisher.GetAccountCreatedEvents(), 0)
	assert.Len(t, eventPublisher.GetDepositCompletedEvents(), 0)
	assert.Len(t, eventPublisher.GetWithdrawalCompletedEvents(), 0)
	assert.Len(t, eventPublisher.GetTransferCompletedEvents(), 0)
	assert.Len(t, eventPublisher.GetTransactionFailedEvents(), 0)
}

// TestFailedOperationNoEvent verifies that failed operations don't publish success events
func TestFailedOperationNoEvent(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()
	eventPublisher := container.GetEventPublisher()

	// Create account with zero balance
	testenv.CreateAccount(t, router, "Henry")

	// Reset to clear account creation event
	eventPublisher.Reset()

	// Attempt withdrawal with insufficient funds
	body := map[string]int{"amount": 1000}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/1/withdraw", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// Verify request failed
	require.Equal(t, http.StatusBadRequest, resp.Code)

	// Verify no withdrawal event was published (since operation failed)
	withdrawalEvents := eventPublisher.GetWithdrawalCompletedEvents()
	assert.Len(t, withdrawalEvents, 0, "Failed withdrawal should not publish WithdrawalCompletedEvent")
}
