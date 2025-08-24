package monitor

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SystemMonitor struct {
	apiURL         string
	isolateMetrics bool
	mu             sync.RWMutex
	stats          *SystemStats
	apiPID         int
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

type SystemStats struct {
	CPUPercent        float64
	MemoryMB          float64
	MemoryPercent     float64
	GoroutineCount    int
	OpenConnections   int
	OpenFiles         int
	CPUSamples        []float64
	MemorySamples     []float64
	MaxCPU            float64
	MaxMemory         float64
	AvgCPU            float64
	AvgMemory         float64
	SystemCPU         float64
	SystemMemory      float64
	TestProcessCPU    float64
	TestProcessMemory float64
	Timestamp         time.Time
}

func NewSystemMonitor(apiURL string, isolateMetrics bool) *SystemMonitor {
	return &SystemMonitor{
		apiURL:         apiURL,
		isolateMetrics: isolateMetrics,
		stats:          &SystemStats{},
		stopChan:       make(chan struct{}),
	}
}

func (m *SystemMonitor) Start(ctx context.Context) error {
	if m.isolateMetrics {
		pid, err := m.findAPIPID()
		if err != nil {
			return fmt.Errorf("failed to find API process: %w", err)
		}
		m.apiPID = pid
	}

	m.wg.Add(1)
	go m.collect(ctx)

	return nil
}

func (m *SystemMonitor) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}

func (m *SystemMonitor) GetStats() *SystemStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	statsCopy := *m.stats
	return &statsCopy
}

func (m *SystemMonitor) collect(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	cpuSamples := make([]float64, 0, 1000)
	memorySamples := make([]float64, 0, 1000)

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			stats := &SystemStats{
				Timestamp: time.Now(),
			}

			if m.isolateMetrics && m.apiPID > 0 {
				cpu, mem, err := m.getProcessStats(m.apiPID)
				if err == nil {
					stats.CPUPercent = cpu
					stats.MemoryMB = mem
					cpuSamples = append(cpuSamples, cpu)
					memorySamples = append(memorySamples, mem)
				}

				testCPU, testMem, err := m.getCurrentProcessStats()
				if err == nil {
					stats.TestProcessCPU = testCPU
					stats.TestProcessMemory = testMem
				}
			} else {
				cpu, mem := m.getSystemStats()
				stats.SystemCPU = cpu
				stats.SystemMemory = mem
			}

			if len(cpuSamples) > 0 {
				stats.MaxCPU = max(cpuSamples)
				stats.AvgCPU = avg(cpuSamples)
			}

			if len(memorySamples) > 0 {
				stats.MaxMemory = max(memorySamples)
				stats.AvgMemory = avg(memorySamples)
			}

			stats.CPUSamples = cpuSamples
			stats.MemorySamples = memorySamples

			m.mu.Lock()
			m.stats = stats
			m.mu.Unlock()
		}
	}
}

func (m *SystemMonitor) findAPIPID() (int, error) {
	portStr := "8080"
	if strings.Contains(m.apiURL, ":") {
		parts := strings.Split(m.apiURL, ":")
		if len(parts) >= 3 {
			portStr = parts[len(parts)-1]
		}
	}

	cmd := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%s", portStr))
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("ps aux | grep 'bank-api\\|main.go' | grep -v grep | awk '{print $2}' | head -1"))
		output, err = cmd.Output()
		if err != nil {
			return 0, fmt.Errorf("failed to find API process: %w", err)
		}
	}

	pidStr := strings.TrimSpace(string(output))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID: %w", err)
	}

	return pid, nil
}

func (m *SystemMonitor) getProcessStats(pid int) (cpu float64, memMB float64, err error) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "%cpu,rss")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("unexpected ps output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected ps fields")
	}

	cpu, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, err
	}

	rssKB, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, err
	}

	memMB = rssKB / 1024.0

	return cpu, memMB, nil
}

func (m *SystemMonitor) getCurrentProcessStats() (cpu float64, memMB float64, err error) {
	cmd := exec.Command("sh", "-c", "ps -p $$ -o %cpu,rss | tail -1")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected ps fields")
	}

	cpu, _ = strconv.ParseFloat(fields[0], 64)
	rssKB, _ := strconv.ParseFloat(fields[1], 64)
	memMB = rssKB / 1024.0

	return cpu, memMB, nil
}

func (m *SystemMonitor) getSystemStats() (cpu float64, memMB float64) {
	cmd := exec.Command("sh", "-c", "top -l 1 | grep 'CPU usage' | awk '{print $3}' | tr -d '%'")
	if output, err := cmd.Output(); err == nil {
		cpu, _ = strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	}

	cmd = exec.Command("sh", "-c", "vm_stat | grep 'Pages active' | awk '{print $3}' | tr -d '.'")
	if output, err := cmd.Output(); err == nil {
		if pages, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64); err == nil {
			memMB = (pages * 4096) / (1024 * 1024)
		}
	}

	return cpu, memMB
}

func (m *SystemMonitor) getNetworkStats(pid int) (connections int, err error) {
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid), "-i")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	tcpCount := 0
	for _, line := range lines {
		if strings.Contains(line, "TCP") && strings.Contains(line, "ESTABLISHED") {
			tcpCount++
		}
	}

	return tcpCount, nil
}

func (m *SystemMonitor) getFileDescriptors(pid int) (int, error) {
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	return len(lines) - 1, nil
}

func (m *SystemMonitor) getGoroutineCount() (int, error) {
	cmd := exec.Command("curl", "-s", fmt.Sprintf("%s/debug/pprof/goroutine?debug=1", m.apiURL))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	re := regexp.MustCompile(`goroutine \d+`)
	matches := re.FindAllString(string(output), -1)
	return len(matches), nil
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

func avg(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}