package handlers

import (
	"bank-api/internal/infrastructure/messaging"
	"bank-api/internal/pkg/errors"
	"bank-api/internal/pkg/logging"
	"bank-api/internal/pkg/telemetry"
	"bank-api/internal/pkg/validation"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func MakeTransferHandler(container HandlerDependencies) gin.HandlerFunc {
	// Extract dependencies once at handler creation time
	db := container.GetDatabase()
	publisher := container.GetEventPublisher()

	return func(c *gin.Context) {
		var req struct {
			FromID int `json:"from"`
			ToID   int `json:"to"`
			Amount int `json:"amount"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			apiErr := errors.NewValidationError("Invalid request format")
			logging.Warn("Invalid JSON in transfer request", map[string]interface{}{
				"error": err.Error(),
				"ip":    c.ClientIP(),
			})
			c.JSON(apiErr.Status, apiErr)
			return
		}

		if err := validation.ValidateAmount(req.Amount); err != nil {
			apiErr := errors.NewInvalidAmountError(err.Error())
			c.JSON(apiErr.Status, apiErr)
			return
		}

		if err := validation.ValidateAccountID(req.FromID); err != nil {
			apiErr := errors.NewValidationError("Invalid from account ID: " + err.Error())
			c.JSON(apiErr.Status, apiErr)
			return
		}

		if err := validation.ValidateAccountID(req.ToID); err != nil {
			apiErr := errors.NewValidationError("Invalid to account ID: " + err.Error())
			c.JSON(apiErr.Status, apiErr)
			return
		}

		if req.FromID == req.ToID {
			apiErr := errors.NewSelfTransferError()
			logging.Warn("Attempted self-transfer", map[string]interface{}{
				"account_id": req.FromID,
				"amount":     req.Amount,
				"ip":         c.ClientIP(),
			})
			c.JSON(apiErr.Status, apiErr)
			return
		}

		// Use atomic transfer operation to prevent race conditions
		from, to, err := db.AtomicTransfer(req.FromID, req.ToID, req.Amount)

		if err != nil {
			// Record failed operation
			metrics.RecordBankingOperation("transfer", "error")

			// Check error type
			if strings.Contains(err.Error(), "insufficient balance") {
				apiErr := errors.NewInsufficientFundsError()
				logging.Warn("Transfer failed: insufficient funds", map[string]interface{}{
					"from_account_id": req.FromID,
					"to_account_id":   req.ToID,
					"amount":          req.Amount,
					"ip":              c.ClientIP(),
				})
				c.JSON(apiErr.Status, apiErr)
			} else {
				apiErr := errors.NewAccountNotFoundError()
				logging.Warn("Transfer failed: account not found", map[string]interface{}{
					"from_account_id": req.FromID,
					"to_account_id":   req.ToID,
					"amount":          req.Amount,
					"error":           err.Error(),
					"ip":              c.ClientIP(),
				})
				c.JSON(apiErr.Status, apiErr)
			}
			return
		}

		// Record successful operation and metrics
		metrics.RecordBankingOperation("transfer", "success")
		metrics.RecordTransferAmount(float64(req.Amount))
		metrics.RecordAccountBalance(float64(from.Balance))
		metrics.RecordAccountBalance(float64(to.Balance))

		// Publish transfer completed event to Kafka
		event := messaging.TransferCompletedEvent{
			FromAccountID:    from.Id,
			ToAccountID:      to.Id,
			Amount:           req.Amount,
			FromBalanceAfter: from.Balance,
			ToBalanceAfter:   to.Balance,
			Timestamp:        time.Now(),
		}
		if err := publisher.PublishTransferCompleted(event); err != nil {
			logging.Error("Failed to publish transfer completed event", err, map[string]interface{}{
				"from_account_id": from.Id,
				"to_account_id":   to.Id,
				"amount":          req.Amount,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Transferência realizada com sucesso",
			"from_balance": from.Balance,
			"to_balance":   to.Balance,
			"from_id":      from.Id,
			"to_id":        to.Id,
			"transferred":  req.Amount,
		})
	}
}
