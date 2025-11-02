package handlers

import (
	"bank-api/internal/domain/account"
	"bank-api/internal/infrastructure/messaging"
	"bank-api/internal/pkg/errors"
	"bank-api/internal/pkg/logging"
	"bank-api/internal/pkg/telemetry"
	"bank-api/internal/pkg/validation"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func MakeCreateAccountHandler(container HandlerDependencies) gin.HandlerFunc {
	// Extract dependencies once at handler creation time
	db := container.GetDatabase()
	publisher := container.GetEventPublisher()

	return func(ctx *gin.Context) {
		var req struct {
			Owner string `json:"owner"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			apiErr := errors.NewValidationError("Invalid request format")
			logging.Warn("Invalid JSON in create account request", map[string]interface{}{
				"error": err.Error(),
				"ip":    ctx.ClientIP(),
			})
			ctx.JSON(apiErr.Status, apiErr)
			return
		}

		if err := validation.ValidateOwnerName(req.Owner); err != nil {
			apiErr := errors.NewValidationError(err.Error())
			logging.Warn("Invalid owner name", map[string]interface{}{
				"owner": req.Owner,
				"error": err.Error(),
				"ip":    ctx.ClientIP(),
			})
			ctx.JSON(apiErr.Status, apiErr)
			return
		}

		id := db.CreateAccount(req.Owner)

		// Record metrics
		metrics.RecordAccountCreation()

		// Publish account created event
		event := messaging.AccountCreatedEvent{
			AccountID: id,
			Owner:     req.Owner,
			Timestamp: time.Now(),
		}
		if err := publisher.PublishAccountCreated(event); err != nil {
			logging.Error("Failed to publish account created event", err, map[string]interface{}{
				"account_id": id,
				"owner":      req.Owner,
			})
			// Don't fail the request if event publishing fails (graceful degradation)
		}

		logging.Info("Account created successfully", map[string]interface{}{
			"account_id": id,
			"owner":      req.Owner,
			"ip":         ctx.ClientIP(),
		})

		ctx.JSON(http.StatusCreated, gin.H{"id": id, "owner": req.Owner})
	}
}

func MakeGetBalanceHandler(container HandlerDependencies) gin.HandlerFunc {
	// Extract dependencies once at handler creation time
	db := container.GetDatabase()

	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			apiErr := errors.NewValidationError("Invalid account ID format")
			logging.Warn("Invalid account ID format", map[string]interface{}{
				"id_param": idStr,
				"error":    err.Error(),
				"ip":       c.ClientIP(),
			})
			c.JSON(apiErr.Status, apiErr)
			return
		}

		if err := validation.ValidateAccountID(id); err != nil {
			apiErr := errors.NewValidationError(err.Error())
			c.JSON(apiErr.Status, apiErr)
			return
		}

		account, ok := db.GetAccount(id)
		if !ok {
			apiErr := errors.NewAccountNotFoundError()
			logging.Warn("Account not found", map[string]interface{}{
				"account_id": id,
				"ip":         c.ClientIP(),
			})
			c.JSON(apiErr.Status, apiErr)
			return
		}

		balance := domain.GetBalance(account)

		// Record balance for distribution metrics
		metrics.RecordAccountBalance(float64(balance))

		logging.Debug("Balance retrieved", map[string]interface{}{
			"account_id": id,
			"balance":    balance,
			"ip":         c.ClientIP(),
		})

		c.JSON(http.StatusOK, gin.H{
			"id":      account.Id,
			"owner":   account.Owner,
			"balance": balance,
		})
	}
}
