package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"github.com/core-banking/perf-test/internal/config"
	"github.com/core-banking/perf-test/internal/generator"
	"github.com/core-banking/perf-test/internal/metrics"
	"github.com/core-banking/perf-test/internal/monitor"
)

type Report struct {
	TestName      string                   `json:"test_name"`
	StartTime     time.Time                `json:"start_time"`
	EndTime       time.Time                `json:"end_time"`
	Duration      time.Duration            `json:"duration"`
	Configuration *config.Config           `json:"configuration"`
	Scenario      *generator.Scenario      `json:"scenario"`
	Performance   *PerformanceMetrics      `json:"performance"`
	System        *SystemMetrics           `json:"system"`
	Prometheus    *metrics.PrometheusMetrics `json:"prometheus,omitempty"`
	Summary       *Summary                 `json:"summary"`
	Errors        []ErrorDetail            `json:"errors,omitempty"`
}

type PerformanceMetrics struct {
	TotalRequests     int64                          `json:"total_requests"`
	SuccessfulRequests int64                          `json:"successful_requests"`
	FailedRequests    int64                          `json:"failed_requests"`
	SuccessRate       float64                        `json:"success_rate"`
	RequestsPerSecond float64                        `json:"requests_per_second"`
	Latency           *LatencyMetrics                `json:"latency"`
	Operations        map[string]*OperationMetrics   `json:"operations"`
}

type LatencyMetrics struct {
	Min    time.Duration `json:"min"`
	Max    time.Duration `json:"max"`
	Mean   time.Duration `json:"mean"`
	Median time.Duration `json:"median"`
	P50    time.Duration `json:"p50"`
	P90    time.Duration `json:"p90"`
	P95    time.Duration `json:"p95"`
	P99    time.Duration `json:"p99"`
	StdDev time.Duration `json:"std_dev"`
}

type OperationMetrics struct {
	Count       int64           `json:"count"`
	SuccessRate float64         `json:"success_rate"`
	MeanLatency time.Duration   `json:"mean_latency"`
	P99Latency  time.Duration   `json:"p99_latency"`
	Errors      map[string]int64 `json:"errors,omitempty"`
}

type SystemMetrics struct {
	API        *ProcessMetrics `json:"api"`
	LoadTester *ProcessMetrics `json:"load_tester,omitempty"`
	Combined   *ProcessMetrics `json:"combined,omitempty"`
}

type ProcessMetrics struct {
	CPUUsage      ResourceUsage `json:"cpu_usage"`
	MemoryUsage   ResourceUsage `json:"memory_usage"`
	Connections   int           `json:"connections,omitempty"`
	FileDescriptors int         `json:"file_descriptors,omitempty"`
	Goroutines    int           `json:"goroutines,omitempty"`
}

type ResourceUsage struct {
	Current float64   `json:"current"`
	Average float64   `json:"average"`
	Max     float64   `json:"max"`
	Samples []float64 `json:"samples,omitempty"`
}

type Summary struct {
	Status         string   `json:"status"`
	TotalOperations int64    `json:"total_operations"`
	Throughput     float64  `json:"throughput_ops_per_sec"`
	P99Latency     string   `json:"p99_latency"`
	ErrorRate      float64  `json:"error_rate"`
	PeakCPU        float64  `json:"peak_cpu_percent"`
	PeakMemory     float64  `json:"peak_memory_mb"`
	Bottlenecks    []string `json:"bottlenecks,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

type ErrorDetail struct {
	Type       string `json:"type"`
	Count      int64  `json:"count"`
	Percentage float64 `json:"percentage"`
	Sample     string `json:"sample,omitempty"`
}

func Generate(stats *metrics.Stats, sysStats *monitor.SystemStats, promStats *metrics.PrometheusMetrics, 
	scenario *generator.Scenario, cfg *config.Config) *Report {
	
	endTime := time.Now()
	startTime := endTime.Add(-stats.Duration)
	
	report := &Report{
		TestName:      scenario.Name,
		StartTime:     startTime,
		EndTime:       endTime,
		Duration:      stats.Duration,
		Configuration: cfg,
		Scenario:      scenario,
		Performance:   generatePerformanceMetrics(stats),
		System:        generateSystemMetrics(sysStats, cfg.IsolateMetrics),
		Prometheus:    promStats,
		Errors:        generateErrorDetails(stats),
	}
	
	report.Summary = generateSummary(report)
	
	return report
}

func generatePerformanceMetrics(stats *metrics.Stats) *PerformanceMetrics {
	perf := &PerformanceMetrics{
		TotalRequests:      stats.TotalRequests,
		SuccessfulRequests: stats.TotalSuccess,
		FailedRequests:     stats.TotalFailures,
		SuccessRate:        stats.SuccessRate,
		RequestsPerSecond:  stats.RequestsPerSecond,
		Latency: &LatencyMetrics{
			Min:    stats.MinLatency,
			Max:    stats.MaxLatency,
			Mean:   stats.MeanLatency,
			Median: stats.MedianLatency,
			P50:    stats.P50Latency,
			P90:    stats.P90Latency,
			P95:    stats.P95Latency,
			P99:    stats.P99Latency,
			StdDev: stats.StdDevLatency,
		},
		Operations: make(map[string]*OperationMetrics),
	}
	
	for opType, opStats := range stats.OperationStats {
		perf.Operations[opType] = &OperationMetrics{
			Count:       opStats.Count,
			SuccessRate: opStats.SuccessRate,
			MeanLatency: opStats.MeanLatency,
			P99Latency:  opStats.P99Latency,
			Errors:      opStats.ErrorDistribution,
		}
	}
	
	return perf
}

func generateSystemMetrics(sysStats *monitor.SystemStats, isolated bool) *SystemMetrics {
	metrics := &SystemMetrics{}
	
	if isolated {
		metrics.API = &ProcessMetrics{
			CPUUsage: ResourceUsage{
				Current: sysStats.CPUPercent,
				Average: sysStats.AvgCPU,
				Max:     sysStats.MaxCPU,
				Samples: sysStats.CPUSamples,
			},
			MemoryUsage: ResourceUsage{
				Current: sysStats.MemoryMB,
				Average: sysStats.AvgMemory,
				Max:     sysStats.MaxMemory,
				Samples: sysStats.MemorySamples,
			},
			Connections:     sysStats.OpenConnections,
			FileDescriptors: sysStats.OpenFiles,
			Goroutines:      sysStats.GoroutineCount,
		}
		
		if sysStats.TestProcessCPU > 0 || sysStats.TestProcessMemory > 0 {
			metrics.LoadTester = &ProcessMetrics{
				CPUUsage: ResourceUsage{
					Current: sysStats.TestProcessCPU,
				},
				MemoryUsage: ResourceUsage{
					Current: sysStats.TestProcessMemory,
				},
			}
		}
	} else {
		metrics.Combined = &ProcessMetrics{
			CPUUsage: ResourceUsage{
				Current: sysStats.SystemCPU,
			},
			MemoryUsage: ResourceUsage{
				Current: sysStats.SystemMemory,
			},
		}
	}
	
	return metrics
}

func generateErrorDetails(stats *metrics.Stats) []ErrorDetail {
	var errors []ErrorDetail
	
	for errType, count := range stats.ErrorDistribution {
		percentage := float64(count) / float64(stats.TotalRequests) * 100
		errors = append(errors, ErrorDetail{
			Type:       errType,
			Count:      count,
			Percentage: percentage,
		})
	}
	
	return errors
}

func generateSummary(report *Report) *Summary {
	summary := &Summary{
		Status:          determineTestStatus(report),
		TotalOperations: report.Performance.TotalRequests,
		Throughput:      report.Performance.RequestsPerSecond,
		P99Latency:      formatDuration(report.Performance.Latency.P99),
		ErrorRate:       (1 - report.Performance.SuccessRate) * 100,
	}
	
	if report.System.API != nil {
		summary.PeakCPU = report.System.API.CPUUsage.Max
		summary.PeakMemory = report.System.API.MemoryUsage.Max
	} else if report.System.Combined != nil {
		summary.PeakCPU = report.System.Combined.CPUUsage.Current
		summary.PeakMemory = report.System.Combined.MemoryUsage.Current
	}
	
	summary.Bottlenecks = identifyBottlenecks(report)
	summary.Recommendations = generateRecommendations(report)
	
	return summary
}

func determineTestStatus(report *Report) string {
	if report.Performance.SuccessRate >= 0.99 && report.Performance.Latency.P99 < 100*time.Millisecond {
		return "EXCELLENT"
	} else if report.Performance.SuccessRate >= 0.95 && report.Performance.Latency.P99 < 500*time.Millisecond {
		return "GOOD"
	} else if report.Performance.SuccessRate >= 0.90 {
		return "ACCEPTABLE"
	}
	return "NEEDS_IMPROVEMENT"
}

func identifyBottlenecks(report *Report) []string {
	var bottlenecks []string
	
	if report.Performance.Latency.P99 > 1*time.Second {
		bottlenecks = append(bottlenecks, "High P99 latency detected")
	}
	
	if report.System.API != nil && report.System.API.CPUUsage.Max > 80 {
		bottlenecks = append(bottlenecks, "CPU usage exceeding 80%")
	}
	
	if report.Performance.SuccessRate < 0.95 {
		bottlenecks = append(bottlenecks, "Error rate above 5%")
	}
	
	for opType, metrics := range report.Performance.Operations {
		if metrics.P99Latency > 2*time.Second {
			bottlenecks = append(bottlenecks, fmt.Sprintf("%s operations showing high latency", opType))
		}
	}
	
	return bottlenecks
}

func generateRecommendations(report *Report) []string {
	var recommendations []string
	
	if report.Performance.Latency.P99 > 1*time.Second {
		recommendations = append(recommendations, "Consider adding caching or optimizing database queries")
	}
	
	if report.System.API != nil && report.System.API.CPUUsage.Max > 80 {
		recommendations = append(recommendations, "Scale horizontally or optimize CPU-intensive operations")
	}
	
	if report.Performance.SuccessRate < 0.99 {
		recommendations = append(recommendations, "Investigate error patterns and improve error handling")
	}
	
	if transferOp, exists := report.Performance.Operations["transfer"]; exists {
		if transferOp.P99Latency > 500*time.Millisecond {
			recommendations = append(recommendations, "Optimize transfer locking mechanism")
		}
	}
	
	return recommendations
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func SaveReport(report *Report, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}
	
	fmt.Printf("\nReport saved to: %s\n", path)
	return nil
}

func PrintSummary(report *Report) {
	fmt.Printf("\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("                    LOAD TEST SUMMARY                          \n")
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("\n")
	fmt.Printf("Test Name:        %s\n", report.TestName)
	fmt.Printf("Duration:         %.2f seconds\n", report.Duration.Seconds())
	fmt.Printf("Status:           %s\n", report.Summary.Status)
	fmt.Printf("\n")
	fmt.Printf("Performance Metrics:\n")
	fmt.Printf("  Total Requests:   %d\n", report.Performance.TotalRequests)
	fmt.Printf("  Success Rate:     %.2f%%\n", report.Performance.SuccessRate*100)
	fmt.Printf("  Throughput:       %.2f ops/sec\n", report.Performance.RequestsPerSecond)
	fmt.Printf("  P99 Latency:      %s\n", report.Summary.P99Latency)
	fmt.Printf("  Mean Latency:     %s\n", formatDuration(report.Performance.Latency.Mean))
	fmt.Printf("\n")
	
	if report.System.API != nil {
		fmt.Printf("System Metrics (API Process):\n")
		fmt.Printf("  Peak CPU:         %.2f%%\n", report.System.API.CPUUsage.Max)
		fmt.Printf("  Avg CPU:          %.2f%%\n", report.System.API.CPUUsage.Average)
		fmt.Printf("  Peak Memory:      %.2f MB\n", report.System.API.MemoryUsage.Max)
		fmt.Printf("  Avg Memory:       %.2f MB\n", report.System.API.MemoryUsage.Average)
		fmt.Printf("\n")
	}
	
	fmt.Printf("Operation Breakdown:\n")
	for opType, metrics := range report.Performance.Operations {
		fmt.Printf("  %s:\n", opType)
		fmt.Printf("    Count:          %d\n", metrics.Count)
		fmt.Printf("    Success Rate:   %.2f%%\n", metrics.SuccessRate*100)
		fmt.Printf("    P99 Latency:    %s\n", formatDuration(metrics.P99Latency))
	}
	
	if len(report.Summary.Bottlenecks) > 0 {
		fmt.Printf("\nBottlenecks Identified:\n")
		for _, bottleneck := range report.Summary.Bottlenecks {
			fmt.Printf("  ⚠ %s\n", bottleneck)
		}
	}
	
	if len(report.Summary.Recommendations) > 0 {
		fmt.Printf("\nRecommendations:\n")
		for _, rec := range report.Summary.Recommendations {
			fmt.Printf("  → %s\n", rec)
		}
	}
	
	fmt.Printf("\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")
}