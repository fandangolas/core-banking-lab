package handlers

import (
	"bank-api/src/db"
	"bank-api/src/logic"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateAccount(ctx *gin.Context) {
	var req struct {
		Owner string `json:"owner"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Owner == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid name"})
		return
	}

	id := db.InMemory.CreateAccount(req.Owner)
	ctx.JSON(http.StatusCreated, gin.H{"id": id, "owner": req.Owner})
}

func GetBalance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid identifier (id)"})
		return
	}

	account, ok := db.InMemory.GetAccount(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	balance := logic.GetBalance(account)
	c.JSON(http.StatusOK, gin.H{
		"id":      account.Id,
		"owner":   account.Owner,
		"balance": balance,
	})
}
