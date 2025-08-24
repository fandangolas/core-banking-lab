package config

import (
	"time"
)

type Config struct {
	APIURL         string
	PrometheusURL  string
	Workers        int
	Duration       time.Duration
	RampUp         time.Duration
	ReportPath     string
	IsolateMetrics bool
}

type TestConfig struct {
	Name               string            `json:"name"`
	TotalOperations    int               `json:"total_operations"`
	AccountCount       int               `json:"account_count"`
	OperationMix       OperationMix      `json:"operation_mix"`
	WorkerConfig       WorkerConfig      `json:"worker_config"`
	TargetMetrics      TargetMetrics     `json:"target_metrics"`
}

type OperationMix struct {
	Deposit  float64 `json:"deposit"`
	Withdraw float64 `json:"withdraw"`
	Transfer float64 `json:"transfer"`
	Balance  float64 `json:"balance"`
}

type WorkerConfig struct {
	MinWorkers      int           `json:"min_workers"`
	MaxWorkers      int           `json:"max_workers"`
	RampUpDuration  time.Duration `json:"ramp_up_duration"`
	RampDownDuration time.Duration `json:"ramp_down_duration"`
	ThinkTime       time.Duration `json:"think_time"`
}

type TargetMetrics struct {
	MaxP99Latency     time.Duration `json:"max_p99_latency"`
	MinSuccessRate    float64       `json:"min_success_rate"`
	TargetRPS         float64       `json:"target_rps"`
}

func (om OperationMix) Validate() bool {
	total := om.Deposit + om.Withdraw + om.Transfer + om.Balance
	return total >= 0.99 && total <= 1.01
}