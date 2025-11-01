package middleware

import (
	"bank-api/internal/pkg/telemetry"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware collects HTTP metrics in Prometheus format
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment in-flight requests
		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get labels
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		statusCode := strconv.Itoa(c.Writer.Status())

		// Record metrics
		metrics.HTTPDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration.Seconds())
		metrics.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()

		// Also record in existing metrics system for compatibility
		metrics.Record(method+" "+endpoint, c.Writer.Status(), duration)
	}
}
