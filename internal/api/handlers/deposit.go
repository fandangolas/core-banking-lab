package handlers

import (
	"bank-api/internal/infrastructure/messaging"
	"bank-api/internal/pkg/idempotency"
	"bank-api/internal/pkg/logging"
	"bank-api/internal/pkg/telemetry"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func MakeDepositHandler(container HandlerDependencies) gin.HandlerFunc {
	// Extract dependencies once at handler creation time
	db := container.GetDatabase()
	publisher := container.GetEventPublisher()

	// Event-driven fire-and-forget pattern:
	// 1. Validate account exists (fail fast)
	// 2. Publish DepositRequestedEvent to Kafka
	// 3. Return 202 Accepted with operation_id for tracking
	// 4. Consumer processes event asynchronously, updates DB, publishes DepositCompletedEvent

	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid identifier (id)"})
			return
		}

		var req struct {
			Amount int `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value"})
			return
		}

		// Fail fast - validate account exists before publishing event
		_, ok := db.GetAccount(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}

		// Generate unique operation ID for tracking (legacy)
		operationID := uuid.New().String()

		// Generate deterministic idempotency key (no DB query!)
		// Same request → same key → consumer deduplicates
		idempotencyKey := idempotency.GenerateKey("deposit", id, req.Amount)

		// Publish deposit request event to Kafka (fire-and-forget)
		event := messaging.DepositRequestedEvent{
			OperationID:    operationID,
			IdempotencyKey: idempotencyKey,
			AccountID:      id,
			Amount:         req.Amount,
			Timestamp:      time.Now(),
		}

		if err := publisher.PublishDepositRequested(event); err != nil {
			logging.Error("Failed to publish deposit request event", err, map[string]interface{}{
				"operation_id": operationID,
				"account_id":   id,
				"amount":       req.Amount,
			})
			metrics.RecordBankingOperation("deposit", "error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process deposit request"})
			return
		}

		// Record successful request acceptance
		metrics.RecordBankingOperation("deposit", "accepted")

		// Return 202 Accepted with operation ID for tracking
		c.JSON(http.StatusAccepted, gin.H{
			"operation_id": operationID,
			"status":       "accepted",
			"message":      "Deposit request accepted and will be processed asynchronously",
		})
	}
}
