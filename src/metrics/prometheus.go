package metrics

import (
	"runtime"
	"time"

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

	// CPU usage
	CPUUsageGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_cpu_usage_seconds_total",
			Help: "Total CPU time consumed by the process in seconds",
		},
	)

	// GC metrics
	GCMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_gc_custom_stats",
			Help: "Custom Go garbage collection statistics",
		},
		[]string{"type"}, // type: pause_total, num_gc, heap_objects
	)

	// Concurrency metrics
	ConcurrencyMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_concurrency_stats",
			Help: "Go concurrency and runtime statistics",
		},
		[]string{"type"}, // type: cgo_calls, num_cpu
	)

	// CPU Core metrics
	CPUCoreMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "banking_cpu_core_stats",
			Help: "CPU core utilization and parallel processing statistics",
		},
		[]string{"type"}, // type: available_cores, max_procs, cpu_efficiency, load_balance
	)

	// CPU metrics
	CPUMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "banking_cpu_stats",
			Help: "Banking application CPU usage and scheduling statistics",
		},
		[]string{"type"}, // type: usage_percent, goroutines_per_cpu, gc_cpu_percent
	)

	// Throttling detection
	ThrottlingMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "banking_throttling_stats",
			Help: "Banking application CPU throttling and pressure statistics",
		},
		[]string{"type"}, // type: potential_throttling, goroutine_pressure, gc_pressure
	)
)

// CPU tracking variables
var (
	lastCPUTime      time.Time
	lastUserTime     time.Duration
	lastSystemTime   time.Duration
	lastRunnableTime time.Duration
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

	// Update GC metrics
	GCMetrics.WithLabelValues("pause_total").Set(float64(m.PauseTotalNs) / 1e9) // Convert to seconds
	GCMetrics.WithLabelValues("num_gc").Set(float64(m.NumGC))
	GCMetrics.WithLabelValues("heap_objects").Set(float64(m.HeapObjects))
	GCMetrics.WithLabelValues("next_gc").Set(float64(m.NextGC))
	GCMetrics.WithLabelValues("gc_cpu_fraction").Set(m.GCCPUFraction * 100) // Convert to percentage

	// Update concurrency metrics
	numCPU := float64(runtime.NumCPU())
	maxProcs := float64(runtime.GOMAXPROCS(0))
	ConcurrencyMetrics.WithLabelValues("num_cpu").Set(numCPU)
	ConcurrencyMetrics.WithLabelValues("num_cgo_call").Set(float64(runtime.NumCgoCall()))
	ConcurrencyMetrics.WithLabelValues("max_procs").Set(maxProcs)

	// Update CPU core metrics
	updateCPUCoreMetrics(numCPU, maxProcs, float64(runtime.NumGoroutine()))

	// Update CPU metrics
	updateCPUMetrics()
}

// updateCPUMetrics collects CPU usage and throttling metrics
func updateCPUMetrics() {
	now := time.Now()

	// Initialize on first run
	if lastCPUTime.IsZero() {
		lastCPUTime = now
		return
	}

	// Calculate time since last measurement
	timeDelta := now.Sub(lastCPUTime).Seconds()
	if timeDelta <= 0 {
		return
	}

	// Get current runtime stats for CPU-related metrics
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	// Estimate CPU usage based on goroutine activity and GC
	// Note: This is an approximation since Go doesn't expose direct CPU usage
	activeGoroutines := float64(runtime.NumGoroutine())
	numCPU := float64(runtime.NumCPU())

	// CPU usage approximation (goroutines per CPU as utilization indicator)
	estimatedCPUUsage := (activeGoroutines / numCPU) * 10 // Scale factor for visibility
	if estimatedCPUUsage > 100 {
		estimatedCPUUsage = 100 // Cap at 100%
	}

	CPUMetrics.WithLabelValues("usage_percent").Set(estimatedCPUUsage)
	CPUMetrics.WithLabelValues("goroutines_per_cpu").Set(activeGoroutines / numCPU)

	// GC CPU usage as percentage
	gcCPUFraction := stats.GCCPUFraction * 100
	CPUMetrics.WithLabelValues("gc_cpu_percent").Set(gcCPUFraction)

	// Throttling detection based on scheduling patterns
	// High number of goroutines relative to CPU cores suggests potential throttling
	if activeGoroutines > numCPU*10 { // Threshold: 10x more goroutines than CPUs
		ThrottlingMetrics.WithLabelValues("potential_throttling").Set(1)
		ThrottlingMetrics.WithLabelValues("goroutine_pressure").Set(activeGoroutines / numCPU)
	} else {
		ThrottlingMetrics.WithLabelValues("potential_throttling").Set(0)
		ThrottlingMetrics.WithLabelValues("goroutine_pressure").Set(activeGoroutines / numCPU)
	}

	// Scheduler pressure indicator
	if stats.NumGC > 0 && gcCPUFraction > 5 { // High GC CPU usage
		ThrottlingMetrics.WithLabelValues("gc_pressure").Set(gcCPUFraction)
	} else {
		ThrottlingMetrics.WithLabelValues("gc_pressure").Set(0)
	}

	lastCPUTime = now
}

// updateCPUCoreMetrics collects CPU core utilization and parallel processing metrics
func updateCPUCoreMetrics(numCPU, maxProcs, goroutines float64) {
	// Available cores
	CPUCoreMetrics.WithLabelValues("available_cores").Set(numCPU)

	// Max processes (GOMAXPROCS)
	CPUCoreMetrics.WithLabelValues("max_procs").Set(maxProcs)

	// Core utilization ratio (how many cores we can actually use vs available)
	coreUtilization := (maxProcs / numCPU) * 100
	CPUCoreMetrics.WithLabelValues("core_utilization_percent").Set(coreUtilization)

	// Parallel efficiency (ideal: goroutines distributed across cores)
	// Values close to maxProcs indicate good parallel utilization
	parallelEfficiency := goroutines / maxProcs
	CPUCoreMetrics.WithLabelValues("parallel_efficiency").Set(parallelEfficiency)

	// CPU pressure indicator (high values suggest core contention)
	if goroutines > maxProcs*5 {
		CPUCoreMetrics.WithLabelValues("core_pressure").Set(1) // High pressure
	} else if goroutines > maxProcs*2 {
		CPUCoreMetrics.WithLabelValues("core_pressure").Set(0.5) // Medium pressure
	} else {
		CPUCoreMetrics.WithLabelValues("core_pressure").Set(0) // Low pressure
	}

	// Load balance score (how evenly work might be distributed)
	// Lower values = better load distribution
	loadImbalance := goroutines / maxProcs
	if loadImbalance > 10 {
		CPUCoreMetrics.WithLabelValues("load_balance_score").Set(0) // Poor balance
	} else if loadImbalance > 5 {
		CPUCoreMetrics.WithLabelValues("load_balance_score").Set(50) // Fair balance
	} else {
		CPUCoreMetrics.WithLabelValues("load_balance_score").Set(100) // Good balance
	}

	// Theoretical max parallel tasks (approximation)
	CPUCoreMetrics.WithLabelValues("max_parallel_capacity").Set(maxProcs * 1000) // Rough estimate
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
