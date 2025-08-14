package handlers

import (
	"bank-api/src/metrics"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var startTime = time.Now()

// PrometheusMetrics exposes metrics in Prometheus format
func PrometheusMetrics(c *gin.Context) {
	// Update system metrics before serving
	updateSystemMetricsForPrometheus()

	// Serve Prometheus metrics
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// updateSystemMetricsForPrometheus updates all system and business metrics
func updateSystemMetricsForPrometheus() {
	// Update system metrics
	metrics.UpdateSystemMetrics()

	// Update uptime
	uptime := time.Since(startTime)
	metrics.UptimeGauge.Set(uptime.Seconds())
}
