package metrics

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics for HTTP requests
var (
	// HTTP request duration histogram
	HTTPDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets, // Default buckets: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// HTTP request total counter
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// HTTP requests currently in flight
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being served",
		},
	)
)

// Prometheus metrics for business operations
var (
	// Account operations
	AccountsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "accounts_created_total",
			Help: "Total number of accounts created",
		},
	)

	// Banking operations
	BankingOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "banking_operations_total",
			Help: "Total number of banking operations",
		},
		[]string{"operation", "status"}, // operation: deposit, withdraw, transfer; status: success, error
	)

	// Transfer amount histogram
	TransferAmountHistogram = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "transfer_amount_centavos",
			Help:    "Distribution of transfer amounts in centavos",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000}, // R$ 1 to R$ 10,000
		},
	)

	// Current account balances distribution
	AccountBalancesHistogram = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "account_balances_centavos",
			Help:    "Distribution of account balances in centavos",
			Buckets: []float64{0, 1000, 5000, 10000, 50000, 100000, 500000, 1000000, 5000000}, // R$ 0 to R$ 50,000
		},
	)

	// Total number of active accounts
	ActiveAccountsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "accounts_active_total",
			Help: "Current number of active accounts in the system",
		},
	)
)

// System metrics
var (
	// Goroutine count
	GoroutinesGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_goroutines_current",
			Help: "Current number of goroutines",
		},
	)

	// Memory usage
	MemoryUsageGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
		[]string{"type"}, // type: heap, stack, sys
	)

	// Application uptime
	UptimeGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "application_uptime_seconds",
			Help: "Application uptime in seconds",
		},
	)
)

// UpdateSystemMetrics updates system-level metrics
func UpdateSystemMetrics() {
	// Update goroutine count
	GoroutinesGauge.Set(float64(runtime.NumGoroutine()))

	// Update memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	MemoryUsageGauge.WithLabelValues("heap").Set(float64(m.HeapInuse))
	MemoryUsageGauge.WithLabelValues("stack").Set(float64(m.StackInuse))
	MemoryUsageGauge.WithLabelValues("sys").Set(float64(m.Sys))
}

// RecordAccountCreation records a new account creation
func RecordAccountCreation() {
	AccountsCreatedTotal.Inc()
	// We'll update active accounts count in the handler
}

// RecordBankingOperation records banking operations (deposit, withdraw, transfer)
func RecordBankingOperation(operation, status string) {
	BankingOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordTransferAmount records the amount of a transfer for distribution analysis
func RecordTransferAmount(amount float64) {
	TransferAmountHistogram.Observe(amount)
}

// RecordAccountBalance records an account balance for distribution analysis
func RecordAccountBalance(balance float64) {
	AccountBalancesHistogram.Observe(balance)
}

// UpdateActiveAccounts updates the count of active accounts
func UpdateActiveAccounts(count float64) {
	ActiveAccountsGauge.Set(count)
}