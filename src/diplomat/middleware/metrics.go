package middleware

import (
	"bank-api/src/metrics"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics is a middleware that records basic request information.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		metrics.Record(c.FullPath(), c.Writer.Status(), time.Since(start))
	}
}
