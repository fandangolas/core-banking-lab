package handlers

import (
	"bank-api/src/diplomat/database"
	"bank-api/src/diplomat/events"
	"bank-api/src/errors"
	"bank-api/src/logging"
	"bank-api/src/models"
	"bank-api/src/validation"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Transfer(c *gin.Context) {
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

	from, ok := database.Repo.GetAccount(req.FromID)
	if !ok {
		apiErr := errors.NewAccountNotFoundError()
		logging.Warn("Source account not found", map[string]interface{}{
			"from_account_id": req.FromID,
			"to_account_id":   req.ToID,
			"amount":          req.Amount,
			"ip":              c.ClientIP(),
		})
		c.JSON(apiErr.Status, apiErr)
		return
	}

	to, ok := database.Repo.GetAccount(req.ToID)
	if !ok {
		apiErr := errors.NewAccountNotFoundError()
		logging.Warn("Destination account not found", map[string]interface{}{
			"from_account_id": req.FromID,
			"to_account_id":   req.ToID,
			"amount":          req.Amount,
			"ip":              c.ClientIP(),
		})
		c.JSON(apiErr.Status, apiErr)
		return
	}

	if from.Id < to.Id {
		from.Mu.Lock()
		to.Mu.Lock()
	} else {
		to.Mu.Lock()
		from.Mu.Lock()
	}
	defer from.Mu.Unlock()
	defer to.Mu.Unlock()

	if from.Balance < req.Amount {
		apiErr := errors.NewInsufficientFundsError()
		logging.Warn("Transfer failed: insufficient funds", map[string]interface{}{
			"from_account_id": req.FromID,
			"to_account_id":   req.ToID,
			"amount":          req.Amount,
			"current_balance": from.Balance,
			"ip":              c.ClientIP(),
		})
		c.JSON(apiErr.Status, apiErr)
		return
	}

	from.Balance -= req.Amount
	to.Balance += req.Amount

	database.Repo.UpdateAccount(from)
	database.Repo.UpdateAccount(to)

	events.BrokerInstance.Publish(models.TransactionEvent{
		Type:        "transfer",
		FromID:      from.Id,
		ToID:        to.Id,
		Amount:      req.Amount,
		FromBalance: from.Balance,
		ToBalance:   to.Balance,
		Timestamp:   time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message":      "Transferência realizada com sucesso",
		"from_balance": from.Balance,
		"to_balance":   to.Balance,
		"from_id":      from.Id,
		"to_id":        to.Id,
		"transferred":  req.Amount,
	})
}
