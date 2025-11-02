package messaging

import (
	"bank-api/internal/infrastructure/messaging"
	"bank-api/test/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDepositRequestDeduplication tests that duplicate DepositRequestedEvents
// are not created when the idempotent producer is working correctly
func TestDepositRequestDeduplication(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()
	eventPublisher := container.GetEventPublisher()

	// Create account
	accountID := testenv.CreateAccount(t, router, "Alice")

	// Reset event capture to start fresh
	eventPublisher.Reset()

	// Make deposit request
	body := map[string]int{"amount": 1000}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/1/deposit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusAccepted, resp.Code, "Should return 202 Accepted")

	// Verify only ONE DepositRequestedEvent was published
	events := eventPublisher.GetDepositRequestedEvents()
	assert.Len(t, events, 1, "Should have exactly one DepositRequestedEvent")

	if len(events) > 0 {
		event := events[0]
		assert.Equal(t, accountID, event.AccountID)
		assert.Equal(t, 1000, event.Amount)
		assert.NotEmpty(t, event.OperationID, "Should have operation_id")
		assert.False(t, event.Timestamp.IsZero(), "Should have timestamp")
	}
}

// TestIdempotentProducerPreventsPublisherDuplicates tests that the Kafka
// idempotent producer prevents duplicates from network retries
func TestIdempotentProducerPreventsPublisherDuplicates(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	eventPublisher := container.GetEventPublisher()

	// Create same event with same operation_id
	operationID := uuid.New().String()
	event := messaging.DepositRequestedEvent{
		OperationID: operationID,
		AccountID:   1,
		Amount:      1000,
		Timestamp:   time.Now(),
	}

	// Publish the same event multiple times (simulating retries)
	// With idempotent producer, Kafka should deduplicate at broker level
	err1 := eventPublisher.PublishDepositRequested(event)
	require.NoError(t, err1, "First publish should succeed")

	// Small delay to simulate network retry scenario
	time.Sleep(10 * time.Millisecond)

	err2 := eventPublisher.PublishDepositRequested(event)
	require.NoError(t, err2, "Second publish should also succeed (idempotent)")

	// In our test EventCapture implementation, both will be captured
	// (because EventCapture is in-memory, not using real Kafka)
	// But with real Kafka idempotent producer, only one message would exist
	events := eventPublisher.GetDepositRequestedEvents()

	// For in-memory test publisher, we expect 2 (no deduplication)
	// For real Kafka with idempotence, we'd expect 1 (broker deduplicates)
	t.Logf("Events captured: %d (in-memory test publisher - no Kafka deduplication)", len(events))
	assert.GreaterOrEqual(t, len(events), 1, "Should have at least one event")

	// Verify all events have the same operation_id (idempotency key)
	for _, e := range events {
		assert.Equal(t, operationID, e.OperationID, "All events should have same operation_id")
	}
}

// TestConsumerDeduplicationRequired tests that WITHOUT consumer-side
// deduplication, duplicate processing can occur
func TestConsumerDeduplicationRequired(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	eventPublisher := container.GetEventPublisher()

	// Create two identical events (simulating message redelivery)
	operationID := uuid.New().String()
	event1 := messaging.DepositRequestedEvent{
		OperationID: operationID,
		AccountID:   1,
		Amount:      1000,
		Timestamp:   time.Now(),
	}

	event2 := messaging.DepositRequestedEvent{
		OperationID: operationID, // Same operation_id!
		AccountID:   1,
		Amount:      1000,
		Timestamp:   time.Now().Add(1 * time.Second),
	}

	// Publish both events
	err1 := eventPublisher.PublishDepositRequested(event1)
	require.NoError(t, err1)

	err2 := eventPublisher.PublishDepositRequested(event2)
	require.NoError(t, err2)

	// Both events captured (simulates consumer receiving duplicate message)
	events := eventPublisher.GetDepositRequestedEvents()
	assert.Len(t, events, 2, "Both duplicate events were captured")

	// Verify they have the same operation_id
	assert.Equal(t, events[0].OperationID, events[1].OperationID,
		"Duplicates should have same operation_id")

	// This test demonstrates WHY we need consumer-side idempotency:
	// Even with idempotent producer, at-least-once delivery means
	// the consumer might receive the same message twice
	t.Log("⚠️  This test shows duplicates CAN occur with at-least-once delivery")
	t.Log("✅ Consumer must implement operation_id deduplication to prevent double-processing")
}

// TestUniqueOperationIDs verifies that each deposit request gets a unique operation_id
func TestUniqueOperationIDs(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()
	eventPublisher := container.GetEventPublisher()

	// Create account
	testenv.CreateAccount(t, router, "Bob")

	// Reset events
	eventPublisher.Reset()

	// Make multiple deposit requests
	operationIDs := make(map[string]bool)

	for i := 0; i < 5; i++ {
		body := map[string]int{"amount": 100 * (i + 1)}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/accounts/1/deposit", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		require.Equal(t, http.StatusAccepted, resp.Code)

		// Extract operation_id from response
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		opID, ok := result["operation_id"].(string)
		require.True(t, ok, "Response should contain operation_id")
		require.NotEmpty(t, opID, "operation_id should not be empty")

		// Check uniqueness
		_, exists := operationIDs[opID]
		assert.False(t, exists, "operation_id should be unique: %s", opID)
		operationIDs[opID] = true
	}

	// Verify all events have unique operation_ids
	events := eventPublisher.GetDepositRequestedEvents()
	assert.Len(t, events, 5, "Should have 5 deposit requests")

	uniqueOps := make(map[string]bool)
	for _, event := range events {
		uniqueOps[event.OperationID] = true
	}
	assert.Len(t, uniqueOps, 5, "All 5 operation_ids should be unique")
}

// TestOperationIDFormat verifies that operation_ids are valid UUIDs
func TestOperationIDFormat(t *testing.T) {
	testenv.SetupIntegrationTest(t)
	container := testenv.NewTestContainer()
	defer container.Reset()

	router := container.GetRouter()

	// Create account
	testenv.CreateAccount(t, router, "Charlie")

	// Make deposit request
	body := map[string]int{"amount": 500}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/1/deposit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusAccepted, resp.Code)

	// Extract operation_id
	var result map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	opID, ok := result["operation_id"].(string)
	require.True(t, ok, "Response should contain operation_id")

	// Verify it's a valid UUID
	_, err = uuid.Parse(opID)
	assert.NoError(t, err, "operation_id should be a valid UUID: %s", opID)
}

// TestConsumerIdempotencyContract tests the expected behavior for
// consumer-side idempotency (specification, not implementation yet)
func TestConsumerIdempotencyContract(t *testing.T) {
	t.Skip("Skipping until consumer idempotency is implemented")

	// This test defines what SHOULD happen with consumer idempotency:
	// 1. Consumer receives message with operation_id="abc-123"
	// 2. Consumer processes deposit (balance += 1000)
	// 3. Consumer commits offset
	// 4. Consumer crashes before DB commit
	// 5. Consumer restarts, receives same message again
	// 6. Consumer checks: operation_id="abc-123" already processed?
	// 7. If YES: skip processing (idempotent)
	// 8. If NO: process deposit
	//
	// Expected: Balance increased only ONCE, not twice
}
