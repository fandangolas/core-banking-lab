package handlers

import (
	"bank-api/src/db"
	"bank-api/src/logic"
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

	account, ok := db.InMemory.GetAccount(id)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conta não encontrada"})
		return
	}

	if err := logic.RemoveAmount(account, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Saldo insuficiente"})
		return
	}

	db.InMemory.UpdateAccount(account)

	balance := logic.GetBalance(account)

	c.JSON(http.StatusOK, gin.H{
		"message": "Saque realizado com sucesso",
		"id":      account.Id,
		"balance": balance,
	})
}
