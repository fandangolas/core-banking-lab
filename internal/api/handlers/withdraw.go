package handlers

import (
	"bank-api/internal/domain/models"
	"bank-api/internal/infrastructure/database"
	"bank-api/internal/infrastructure/events"
	"bank-api/internal/infrastructure/messaging"
	"bank-api/internal/pkg/logging"
	"bank-api/internal/pkg/telemetry"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func MakeWithdrawHandler(container HandlerDependencies) gin.HandlerFunc {
	// Extract dependencies once at handler creation time
	db := container.GetDatabase()
	publisher := container.GetEventPublisher()

	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
			return
		}

		var req struct {
			Amount int `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Valor inválido"})
			return
		}

		// Use atomic withdraw operation to prevent race conditions
		account, err := db.AtomicWithdraw(id, req.Amount)

		if err != nil {
			// Record failed operation
			metrics.RecordBankingOperation("withdraw", "error")

			// Check if account not found or insufficient balance
			if strings.Contains(err.Error(), "account not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Conta não encontrada"})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Saldo insuficiente"})
			}
			return
		}

		balance := account.Balance

		// Record successful operation and metrics
		metrics.RecordBankingOperation("withdraw", "success")
		metrics.RecordAccountBalance(float64(balance))

		// Publish legacy event (for backward compatibility)
		events.GetBroker().Publish(models.TransactionEvent{
			Type:      "withdraw",
			AccountID: account.Id,
			Amount:    req.Amount,
			Balance:   balance,
			Timestamp: time.Now(),
		})

		// Publish withdrawal completed event to Kafka
		event := messaging.WithdrawalCompletedEvent{
			AccountID:    account.Id,
			Amount:       req.Amount,
			BalanceAfter: balance,
			Timestamp:    time.Now(),
		}
		if err := publisher.PublishWithdrawalCompleted(event); err != nil {
			logging.Error("Failed to publish withdrawal completed event", err, map[string]interface{}{
				"account_id": account.Id,
				"amount":     req.Amount,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Saque realizado com sucesso",
			"id":      account.Id,
			"balance": balance,
		})
	}
}

// Legacy function for backward compatibility - can be removed after migration
func Withdraw(c *gin.Context) {
	MakeWithdrawHandler(&simpleContainer{
		db:        database.Repo,
		publisher: messaging.NewNoOpEventPublisher(),
	})(c)
}
