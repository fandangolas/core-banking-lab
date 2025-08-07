package handlers

import (
	"bank-api/src/diplomat/database"
	"bank-api/src/domain"
	"net/http"
	"strconv"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Saldo insuficiente"})
		return
	}

	database.Repo.UpdateAccount(account)

	balance := domain.GetBalance(account)

	c.JSON(http.StatusOK, gin.H{
		"message": "Saque realizado com sucesso",
		"id":      account.Id,
		"balance": balance,
	})
}
