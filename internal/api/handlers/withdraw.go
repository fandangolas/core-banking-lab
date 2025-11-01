package handlers

import (
	"bank-api/internal/domain/account"
	"bank-api/internal/domain/models"
	"bank-api/internal/infrastructure/database"
	"bank-api/internal/infrastructure/events"
	"bank-api/internal/pkg/telemetry"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Withdraw(c *gin.Context) {
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

	account, ok := database.Repo.GetAccount(id)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conta não encontrada"})
		return
	}

	if err := domain.RemoveAmount(account, req.Amount); err != nil {
		// Record failed operation
		metrics.RecordBankingOperation("withdraw", "error")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Saldo insuficiente"})
		return
	}

	database.Repo.UpdateAccount(account)

	balance := domain.GetBalance(account)

	// Record successful operation and metrics
	metrics.RecordBankingOperation("withdraw", "success")
	metrics.RecordAccountBalance(float64(balance))

	events.GetBroker().Publish(models.TransactionEvent{
		Type:      "withdraw",
		AccountID: account.Id,
		Amount:    req.Amount,
		Balance:   balance,
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Saque realizado com sucesso",
		"id":      account.Id,
		"balance": balance,
	})
}
