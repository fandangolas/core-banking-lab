package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	
	"github.com/core-banking/perf-test/internal/config"
	"github.com/core-banking/perf-test/internal/generator"
	"github.com/core-banking/perf-test/internal/metrics"
	"github.com/core-banking/perf-test/internal/monitor"
	"github.com/core-banking/perf-test/internal/reporter"
	"github.com/core-banking/perf-test/internal/server"
)

func main() {
	var (
		apiURL          = flag.String("api-url", "http://localhost:8080", "Core Banking API URL")
		prometheusURL   = flag.String("prometheus-url", "http://localhost:9090", "Prometheus server URL")
		mode            = flag.String("mode", "cli", "Run mode: cli or server")
		serverPort      = flag.Int("server-port", 9999, "Load test server port")
		workers         = flag.Int("workers", 100, "Number of concurrent workers")
		duration        = flag.Duration("duration", 60*time.Second, "Test duration")
		rampUp          = flag.Duration("ramp-up", 10*time.Second, "Ramp-up period")
		scenarioFile    = flag.String("scenario", "", "Path to scenario file")
		reportPath      = flag.String("report", "./reports", "Path to save reports")
		isolateMetrics  = flag.Bool("isolate", true, "Isolate API metrics from test metrics")
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	cfg := &config.Config{
		APIURL:         *apiURL,
		PrometheusURL:  *prometheusURL,
		Workers:        *workers,
		Duration:       *duration,
		RampUp:         *rampUp,
		ReportPath:     *reportPath,
		IsolateMetrics: *isolateMetrics,
	}

	if *mode == "server" {
		runServer(ctx, cfg, *serverPort)
	} else {
		runCLI(ctx, cfg, *scenarioFile)
	}
}

func runServer(ctx context.Context, cfg *config.Config, port int) {
	srv := server.New(cfg, port)
	log.Printf("Starting load test server on port %d", port)
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runCLI(ctx context.Context, cfg *config.Config, scenarioFile string) {
	var scenario *generator.Scenario
	var err error

	if scenarioFile != "" {
		scenario, err = generator.LoadScenario(scenarioFile)
		if err != nil {
			log.Fatalf("Failed to load scenario: %v", err)
		}
	} else {
		scenario = generator.DefaultScenario()
	}

	log.Printf("Starting load test with %d workers for %v", cfg.Workers, cfg.Duration)
	log.Printf("Scenario: %s", scenario.Name)

	collector := metrics.NewCollector()
	systemMonitor := monitor.NewSystemMonitor(cfg.APIURL, cfg.IsolateMetrics)
	prometheusCollector := metrics.NewPrometheusCollector(cfg.PrometheusURL)

	if err := systemMonitor.Start(ctx); err != nil {
		log.Fatalf("Failed to start system monitor: %v", err)
	}

	gen := generator.New(cfg, scenario, collector)
	
	testCtx, testCancel := context.WithTimeout(ctx, cfg.Duration)
	defer testCancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		gen.Run(testCtx)
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				stats := collector.GetStats()
				sysStats := systemMonitor.GetStats()
				fmt.Printf("\n=== Live Stats ===\n")
				fmt.Printf("Requests: %d | Success: %.2f%% | P99: %.2fms\n",
					stats.TotalRequests,
					stats.SuccessRate*100,
					stats.P99Latency.Seconds()*1000)
				fmt.Printf("API CPU: %.2f%% | API Memory: %.2f MB | RPS: %.2f\n",
					sysStats.CPUPercent,
					sysStats.MemoryMB,
					stats.RequestsPerSecond)
			case <-testCtx.Done():
				return
			}
		}
	}()

	wg.Wait()

	finalStats := collector.GetStats()
	finalSysStats := systemMonitor.GetStats()
	promStats, _ := prometheusCollector.Collect(ctx, cfg.Duration)

	report := reporter.Generate(finalStats, finalSysStats, promStats, scenario, cfg)
	
	reportFile := fmt.Sprintf("%s/report_%d.json", cfg.ReportPath, time.Now().Unix())
	if err := reporter.SaveReport(report, reportFile); err != nil {
		log.Printf("Failed to save report: %v", err)
	}

	reporter.PrintSummary(report)
}