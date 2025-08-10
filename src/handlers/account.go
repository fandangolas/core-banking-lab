package handlers

import (
	"bank-api/src/diplomat/database"
	"bank-api/src/domain"
	"bank-api/src/errors"
	"bank-api/src/logging"
	"bank-api/src/validation"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateAccount(ctx *gin.Context) {
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

	id := database.Repo.CreateAccount(req.Owner)
	
	logging.Info("Account created successfully", map[string]interface{}{
		"account_id": id,
		"owner":      req.Owner,
		"ip":         ctx.ClientIP(),
	})
	
	ctx.JSON(http.StatusCreated, gin.H{"id": id, "owner": req.Owner})
}

func GetBalance(c *gin.Context) {
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

	account, ok := database.Repo.GetAccount(id)
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
