package handlers

import (
	"bank-api/src/diplomat/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Transfer(c *gin.Context) {
	var req struct {
		FromID int `json:"from"`
		ToID   int `json:"to"`
		Amount int `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	if req.FromID == req.ToID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Não é possível transferir para a mesma conta"})
		return
	}

	from, ok := database.Repo.GetAccount(req.FromID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conta de origem não encontrada"})
		return
	}

	to, ok := database.Repo.GetAccount(req.ToID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conta de destino não encontrada"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Saldo insuficiente na conta de origem"})
		return
	}

	from.Balance -= req.Amount
	to.Balance += req.Amount

	database.Repo.UpdateAccount(from)
	database.Repo.UpdateAccount(to)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Transferência realizada com sucesso",
		"from_balance": from.Balance,
		"to_balance":   to.Balance,
		"from_id":      from.Id,
		"to_id":        to.Id,
		"transferred":  req.Amount,
	})
}
