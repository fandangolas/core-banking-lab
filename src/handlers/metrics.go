package handlers

import (
	"bank-api/src/metrics"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMetrics returns the collected request metrics as JSON.
func GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, metrics.List())
}
