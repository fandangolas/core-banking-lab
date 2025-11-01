package handlers

import (
	"bank-api/internal/pkg/telemetry"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMetrics returns the collected request metrics as JSON.
func GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, metrics.List())
}
