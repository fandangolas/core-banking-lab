package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	
	"github.com/core-banking/perf-test/internal/config"
	"github.com/core-banking/perf-test/internal/generator"
	"github.com/core-banking/perf-test/internal/metrics"
	"github.com/core-banking/perf-test/internal/monitor"
	"github.com/core-banking/perf-test/internal/reporter"
)

type Server struct {
	config       *config.Config
	port         int
	router       *mux.Router
	upgrader     websocket.Upgrader
	mu           sync.RWMutex
	activeTest   *ActiveTest
	testHistory  []*reporter.Report
	wsClients    map[*websocket.Conn]bool
	wsClientsMu  sync.RWMutex
}

type ActiveTest struct {
	ID            string
	Status        string
	StartTime     time.Time
	Scenario      *generator.Scenario
	Collector     *metrics.Collector
	Monitor       *monitor.SystemMonitor
	Generator     *generator.Generator
	Cancel        context.CancelFunc
	LiveStats     *LiveStats
}

type LiveStats struct {
	Timestamp         time.Time       `json:"timestamp"`
	TotalRequests     int64           `json:"total_requests"`
	SuccessRate       float64         `json:"success_rate"`
	RequestsPerSecond float64         `json:"requests_per_second"`
	P99Latency        float64         `json:"p99_latency_ms"`
	CPUUsage          float64         `json:"cpu_usage"`
	MemoryUsage       float64         `json:"memory_usage"`
	Operations        map[string]*OpStats `json:"operations"`
}

type OpStats struct {
	Count       int64   `json:"count"`
	SuccessRate float64 `json:"success_rate"`
	P99Latency  float64 `json:"p99_latency_ms"`
}

type TestRequest struct {
	Name            string                     `json:"name"`
	TotalOperations int                        `json:"total_operations"`
	AccountCount    int                        `json:"account_count"`
	Workers         int                        `json:"workers"`
	Duration        int                        `json:"duration_seconds"`
	RampUp          int                        `json:"ramp_up_seconds"`
	ThinkTimeMs     int                        `json:"think_time_ms"`
	OperationMix    map[string]float64         `json:"operation_mix"`
	AmountRange     struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	}                                          `json:"amount_range"`
}

func New(cfg *config.Config, port int) *Server {
	s := &Server{
		config:      cfg,
		port:        port,
		router:      mux.NewRouter(),
		upgrader:    websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		testHistory: make([]*reporter.Report, 0),
		wsClients:   make(map[*websocket.Conn]bool),
	}
	
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/", s.handleIndex).Methods("GET")
	s.router.HandleFunc("/favicon.ico", s.handleFavicon).Methods("GET")
	s.router.HandleFunc("/api/test/start", s.handleStartTest).Methods("POST")
	s.router.HandleFunc("/api/test/stop", s.handleStopTest).Methods("POST")
	s.router.HandleFunc("/api/test/status", s.handleTestStatus).Methods("GET")
	s.router.HandleFunc("/api/test/history", s.handleTestHistory).Methods("GET")
	s.router.HandleFunc("/api/test/report/{id}", s.handleGetReport).Methods("GET")
	s.router.HandleFunc("/api/scenarios", s.handleGetScenarios).Methods("GET")
	s.router.HandleFunc("/ws/stats", s.handleWebSocket)
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
}

func (s *Server) Start(ctx context.Context) error {
	go s.broadcastStats(ctx)
	
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}
	
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	
	log.Printf("Load test server listening on http://localhost:%d", s.port)
	return srv.ListenAndServe()
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func (s *Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStartTest(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.activeTest != nil && (s.activeTest.Status == "running" || s.activeTest.Status == "completed") {
		if s.activeTest.Status == "completed" {
			// Clean up completed test before starting new one
			log.Printf("Cleaning up completed test %s", s.activeTest.ID)
			s.activeTest = nil
		} else {
			http.Error(w, "Test already running", http.StatusConflict)
			return
		}
	}
	
	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	scenario := &generator.Scenario{
		Name:             req.Name,
		Description:      fmt.Sprintf("Test with %d operations on %d accounts", req.TotalOperations, req.AccountCount),
		Accounts:         req.AccountCount,
		TargetOperations: int64(req.TotalOperations),
		Distribution: map[generator.OperationType]float64{
			generator.OpDeposit:  req.OperationMix["deposit"],
			generator.OpWithdraw: req.OperationMix["withdraw"],
			generator.OpTransfer: req.OperationMix["transfer"],
			generator.OpBalance:  req.OperationMix["balance"],
		},
		InitialBalance: 10000.00,
		MinAmount:      req.AmountRange.Min,
		MaxAmount:      req.AmountRange.Max,
		ThinkTime:      time.Duration(req.ThinkTimeMs) * time.Millisecond,
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	testCfg := *s.config
	testCfg.Workers = req.Workers
	testCfg.Duration = time.Duration(req.Duration) * time.Second
	testCfg.RampUp = time.Duration(req.RampUp) * time.Second
	
	collector := metrics.NewCollector()
	systemMonitor := monitor.NewSystemMonitor(testCfg.APIURL, testCfg.IsolateMetrics)
	gen := generator.New(&testCfg, scenario, collector)
	
	s.activeTest = &ActiveTest{
		ID:        fmt.Sprintf("test-%d", time.Now().Unix()),
		Status:    "running",
		StartTime: time.Now(),
		Scenario:  scenario,
		Collector: collector,
		Monitor:   systemMonitor,
		Generator: gen,
		Cancel:    cancel,
	}
	
	go s.runTest(ctx, &testCfg)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     s.activeTest.ID,
		"status": "started",
	})
}

func (s *Server) runTest(ctx context.Context, cfg *config.Config) {
	test := s.activeTest
	
	if err := test.Monitor.Start(ctx); err != nil {
		log.Printf("Failed to start system monitor: %v", err)
	}
	
	// No timeout - test will stop when target operations are reached
	log.Printf("Starting test execution for %s", test.ID)
	test.Generator.Run(ctx)
	
	s.mu.Lock()
	test.Status = "completed"
	log.Printf("Test %s completed, status set to completed", test.ID)
	s.mu.Unlock()
	
	finalStats := test.Collector.GetStats()
	finalSysStats := test.Monitor.GetStats()
	
	prometheusCollector := metrics.NewPrometheusCollector(cfg.PrometheusURL)
	promStats, _ := prometheusCollector.Collect(context.Background(), cfg.Duration)
	
	report := reporter.Generate(finalStats, finalSysStats, promStats, test.Scenario, cfg)
	
	s.mu.Lock()
	s.testHistory = append(s.testHistory, report)
	s.activeTest = nil
	s.mu.Unlock()
	
	reportFile := fmt.Sprintf("%s/report_%s.json", cfg.ReportPath, test.ID)
	reporter.SaveReport(report, reportFile)
}

func (s *Server) handleStopTest(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.activeTest == nil || s.activeTest.Status != "running" {
		// Return success even if no test is running - this prevents UI errors
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "no_test_running"})
		return
	}
	
	s.activeTest.Cancel()
	s.activeTest.Status = "stopped"
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (s *Server) handleTestStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Clean up completed tests automatically
	if s.activeTest != nil && s.activeTest.Status == "completed" {
		log.Printf("Auto-cleaning completed test %s", s.activeTest.ID)
		s.activeTest = nil
	}
	
	status := map[string]interface{}{
		"running": s.activeTest != nil && s.activeTest.Status == "running",
	}
	
	if s.activeTest != nil {
		status["test_id"] = s.activeTest.ID
		status["status"] = s.activeTest.Status
		status["start_time"] = s.activeTest.StartTime
		status["scenario"] = s.activeTest.Scenario.Name
		
		if s.activeTest.Status == "running" {
			stats := s.activeTest.Collector.GetStats()
			sysStats := s.activeTest.Monitor.GetStats()
			
			status["live_stats"] = &LiveStats{
				Timestamp:         time.Now(),
				TotalRequests:     stats.TotalRequests,
				SuccessRate:       stats.SuccessRate,
				RequestsPerSecond: stats.RequestsPerSecond,
				P99Latency:        float64(stats.P99Latency.Milliseconds()),
				CPUUsage:          sysStats.CPUPercent,
				MemoryUsage:       sysStats.MemoryMB,
				Operations:        make(map[string]*OpStats),
			}
			
			for opType, opStat := range stats.OperationStats {
				status["live_stats"].(*LiveStats).Operations[opType] = &OpStats{
					Count:       opStat.Count,
					SuccessRate: opStat.SuccessRate,
					P99Latency:  float64(opStat.P99Latency.Milliseconds()),
				}
			}
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleTestHistory(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	history := make([]map[string]interface{}, 0, len(s.testHistory))
	for _, report := range s.testHistory {
		history = append(history, map[string]interface{}{
			"id":         report.TestName,
			"start_time": report.StartTime,
			"duration":   report.Duration.Seconds(),
			"status":     report.Summary.Status,
			"throughput": report.Performance.RequestsPerSecond,
			"p99_latency": report.Performance.Latency.P99.Milliseconds(),
			"success_rate": report.Performance.SuccessRate,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, report := range s.testHistory {
		if report.TestName == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(report)
			return
		}
	}
	
	http.Error(w, "Report not found", http.StatusNotFound)
}

func (s *Server) handleGetScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios := []map[string]interface{}{
		{
			"name":        "Default",
			"description": "Balanced mix of operations",
			"preset":      generator.DefaultScenario(),
		},
		{
			"name":        "High Concurrency",
			"description": "Heavy transfer load",
			"preset":      generator.HighConcurrencyScenario(),
		},
		{
			"name":        "Read Heavy",
			"description": "Mostly balance checks",
			"preset":      generator.ReadHeavyScenario(),
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scenarios)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()
	
	s.wsClientsMu.Lock()
	s.wsClients[conn] = true
	s.wsClientsMu.Unlock()
	
	defer func() {
		s.wsClientsMu.Lock()
		delete(s.wsClients, conn)
		s.wsClientsMu.Unlock()
	}()
	
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) broadcastStats(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			if s.activeTest != nil && s.activeTest.Status == "running" {
				stats := s.activeTest.Collector.GetStats()
				sysStats := s.activeTest.Monitor.GetStats()
				
				liveStats := &LiveStats{
					Timestamp:         time.Now(),
					TotalRequests:     stats.TotalRequests,
					SuccessRate:       stats.SuccessRate,
					RequestsPerSecond: stats.RequestsPerSecond,
					P99Latency:        float64(stats.P99Latency.Milliseconds()),
					CPUUsage:          sysStats.CPUPercent,
					MemoryUsage:       sysStats.MemoryMB,
					Operations:        make(map[string]*OpStats),
				}
				
				for opType, opStat := range stats.OperationStats {
					liveStats.Operations[opType] = &OpStats{
						Count:       opStat.Count,
						SuccessRate: opStat.SuccessRate,
						P99Latency:  float64(opStat.P99Latency.Milliseconds()),
					}
				}
				
				s.activeTest.LiveStats = liveStats
				s.mu.RUnlock()
				
				s.wsClientsMu.RLock()
				for client := range s.wsClients {
					if err := client.WriteJSON(liveStats); err != nil {
						client.Close()
					}
				}
				s.wsClientsMu.RUnlock()
			} else {
				s.mu.RUnlock()
			}
		}
	}
}