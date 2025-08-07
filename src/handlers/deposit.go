package handlers

import (
	"bank-api/src/db"
	"bank-api/src/logic"
	"net/http"
	"strconv"

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

	account, ok := db.InMemory.GetAccount(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	if err := logic.AddAmount(account, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	balance := logic.GetBalance(account)

	c.JSON(http.StatusOK, gin.H{
		"id":      account.Id,
		"balance": balance,
	})
}
