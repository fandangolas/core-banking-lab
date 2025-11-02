package handlers

import (
	"bank-api/internal/domain/account"
	"bank-api/internal/domain/models"
	"bank-api/internal/infrastructure/database"
	"bank-api/internal/infrastructure/events"
	"bank-api/internal/infrastructure/messaging"
	"bank-api/internal/pkg/logging"
	"bank-api/internal/pkg/telemetry"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Deposit(c *gin.Context) {
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

	account, ok := database.Repo.GetAccount(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	if err := domain.AddAmount(account, req.Amount); err != nil {
		// Record failed operation
		metrics.RecordBankingOperation("deposit", "error")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.Repo.UpdateAccount(account)

	balance := domain.GetBalance(account)

	// Record successful operation and metrics
	metrics.RecordBankingOperation("deposit", "success")
	metrics.RecordAccountBalance(float64(balance))

	// Publish legacy event (for backward compatibility)
	events.GetBroker().Publish(models.TransactionEvent{
		Type:      "deposit",
		AccountID: account.Id,
		Amount:    req.Amount,
		Balance:   balance,
		Timestamp: time.Now(),
	})

	// Publish deposit completed event to Kafka
	publisher := GetEventPublisher(c)
	event := messaging.DepositCompletedEvent{
		AccountID:    account.Id,
		Amount:       req.Amount,
		BalanceAfter: balance,
		Timestamp:    time.Now(),
	}
	if err := publisher.PublishDepositCompleted(event); err != nil {
		logging.Error("Failed to publish deposit completed event", err, map[string]interface{}{
			"account_id": account.Id,
			"amount":     req.Amount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      account.Id,
		"balance": balance,
	})
}
