package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"bank-api/internal/pkg/logging"
	metrics "bank-api/internal/pkg/telemetry"

	"github.com/IBM/sarama"
)

// AsyncProducer wraps Kafka async producer with comprehensive error monitoring
type AsyncProducer struct {
	producer sarama.AsyncProducer
	config   *Config

	// Error monitoring
	errorCount   atomic.Int64
	successCount atomic.Int64
	droppedCount atomic.Int64

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
	closed bool

	// Metrics reporting
	lastReportTime time.Time
	reportInterval time.Duration
}

// ProducerMetrics holds current producer statistics
type ProducerMetrics struct {
	SuccessCount int64
	ErrorCount   int64
	DroppedCount int64
	ErrorRate    float64
	Throughput   float64
}

// NewAsyncProducer creates a new high-performance async Kafka producer
func NewAsyncProducer(config *Config) (*AsyncProducer, error) {
	saramaConfig, err := config.ToSaramaConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama config: %w", err)
	}

	// Enable error returns to monitor failures
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Return.Successes = false // Disable success tracking for performance

	// Maximum throughput configuration
	saramaConfig.Producer.RequiredAcks = sarama.NoResponse       // Fire-and-forget
	saramaConfig.Producer.Compression = sarama.CompressionSnappy  // Compress for efficiency
	saramaConfig.Producer.Flush.Frequency = 10 * time.Millisecond
	saramaConfig.Producer.Flush.Messages = 1000
	saramaConfig.Producer.Flush.MaxMessages = 10000
	saramaConfig.ChannelBufferSize = 500000 // Massive buffer
	saramaConfig.Net.MaxOpenRequests = 100  // High parallelism

	producer, err := sarama.NewAsyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create async kafka producer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	ap := &AsyncProducer{
		producer:       producer,
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
		lastReportTime: time.Now(),
		reportInterval: 30 * time.Second, // Report metrics every 30s
	}

	// Start error monitoring goroutine
	ap.wg.Add(1)
	go ap.monitorErrors()

	// Start metrics reporting goroutine
	ap.wg.Add(1)
	go ap.reportMetrics()

	logging.Info("Async Kafka producer initialized", map[string]interface{}{
		"brokers":         config.Brokers,
		"client_id":       config.ClientID,
		"buffer_size":     500000,
		"compression":     "snappy",
		"required_acks":   "none",
		"max_open_reqs":   100,
		"flush_frequency": "10ms",
		"flush_messages":  1000,
	})

	return ap, nil
}

// PublishEventAsync publishes an event asynchronously (non-blocking)
func (ap *AsyncProducer) PublishEventAsync(topic string, key string, event interface{}) error {
	ap.mu.RLock()
	if ap.closed {
		ap.mu.RUnlock()
		ap.droppedCount.Add(1)
		logging.Warn("Event dropped - producer closed", map[string]interface{}{
			"topic": topic,
			"key":   key,
		})
		return fmt.Errorf("producer is closed")
	}
	ap.mu.RUnlock()

	// Serialize event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		ap.droppedCount.Add(1)
		logging.Error("Failed to marshal event", err, map[string]interface{}{
			"topic": topic,
			"key":   key,
		})
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(eventJSON),
	}

	// Send to producer input channel (non-blocking with timeout)
	select {
	case ap.producer.Input() <- msg:
		// Message queued successfully
		return nil
	case <-time.After(100 * time.Millisecond):
		// Queue is full, drop the message
		ap.droppedCount.Add(1)

		logging.Warn("Event dropped - producer queue full", map[string]interface{}{
			"topic":         topic,
			"key":           key,
			"dropped_total": ap.droppedCount.Load(),
		})

		// Record metric
		metrics.RecordEventDropped("queue_full")

		return fmt.Errorf("producer queue full - event dropped")
	case <-ap.ctx.Done():
		ap.droppedCount.Add(1)
		return fmt.Errorf("producer shutting down")
	}
}

// monitorErrors monitors the producer error channel
func (ap *AsyncProducer) monitorErrors() {
	defer ap.wg.Done()

	for {
		select {
		case err := <-ap.producer.Errors():
			if err == nil {
				continue
			}

			ap.errorCount.Add(1)

			// Log error with details
			logging.Error("Kafka producer error", err.Err, map[string]interface{}{
				"topic":       err.Msg.Topic,
				"key":         string(err.Msg.Key.(sarama.StringEncoder)),
				"error_count": ap.errorCount.Load(),
			})

			// Record metric
			metrics.RecordEventPublishingError("kafka_error")

		case <-ap.ctx.Done():
			return
		}
	}
}

// reportMetrics periodically reports producer metrics
func (ap *AsyncProducer) reportMetrics() {
	defer ap.wg.Done()

	ticker := time.NewTicker(ap.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := ap.GetMetrics()

			// Log metrics summary
			logging.Info("Kafka producer metrics", map[string]interface{}{
				"success_count": metrics.SuccessCount,
				"error_count":   metrics.ErrorCount,
				"dropped_count": metrics.DroppedCount,
				"error_rate":    fmt.Sprintf("%.2f%%", metrics.ErrorRate),
				"throughput":    fmt.Sprintf("%.2f msg/s", metrics.Throughput),
			})

			// Alert if error rate is high
			if metrics.ErrorRate > 10.0 {
				logging.Warn("High Kafka producer error rate detected!", map[string]interface{}{
					"error_rate":  fmt.Sprintf("%.2f%%", metrics.ErrorRate),
					"error_count": metrics.ErrorCount,
					"action":      "investigate Kafka connectivity",
				})
			}

			// Alert if messages being dropped
			if metrics.DroppedCount > 0 {
				logging.Warn("Kafka producer dropping messages!", map[string]interface{}{
					"dropped_count": metrics.DroppedCount,
					"action":        "system overloaded or Kafka down",
				})
			}

		case <-ap.ctx.Done():
			return
		}
	}
}

// GetMetrics returns current producer metrics
func (ap *AsyncProducer) GetMetrics() ProducerMetrics {
	successCount := ap.successCount.Load()
	errorCount := ap.errorCount.Load()
	droppedCount := ap.droppedCount.Load()

	total := successCount + errorCount
	errorRate := 0.0
	if total > 0 {
		errorRate = (float64(errorCount) / float64(total)) * 100.0
	}

	// Calculate throughput since last report
	now := time.Now()
	duration := now.Sub(ap.lastReportTime).Seconds()
	throughput := 0.0
	if duration > 0 {
		throughput = float64(successCount) / duration
	}

	return ProducerMetrics{
		SuccessCount: successCount,
		ErrorCount:   errorCount,
		DroppedCount: droppedCount,
		ErrorRate:    errorRate,
		Throughput:   throughput,
	}
}

// IncrementSuccess manually increments success count (for external tracking)
func (ap *AsyncProducer) IncrementSuccess() {
	ap.successCount.Add(1)
}

// Close gracefully shuts down the producer
func (ap *AsyncProducer) Close() error {
	ap.mu.Lock()
	if ap.closed {
		ap.mu.Unlock()
		return nil
	}
	ap.closed = true
	ap.mu.Unlock()

	logging.Info("Closing async Kafka producer...", nil)

	// Stop accepting new messages
	ap.cancel()

	// Close producer (waits for pending messages)
	closeErr := ap.producer.Close()

	// Wait for monitoring goroutines to finish
	done := make(chan struct{})
	go func() {
		ap.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logging.Info("Async Kafka producer closed gracefully", nil)
	case <-time.After(30 * time.Second):
		logging.Warn("Async Kafka producer shutdown timeout", nil)
	}

	// Final metrics report
	finalMetrics := ap.GetMetrics()
	logging.Info("Final Kafka producer metrics", map[string]interface{}{
		"total_success": finalMetrics.SuccessCount,
		"total_errors":  finalMetrics.ErrorCount,
		"total_dropped": finalMetrics.DroppedCount,
		"error_rate":    fmt.Sprintf("%.2f%%", finalMetrics.ErrorRate),
	})

	return closeErr
}

// IsHealthy checks if the producer is healthy
func (ap *AsyncProducer) IsHealthy() bool {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	if ap.closed {
		return false
	}

	// Consider unhealthy if error rate is very high
	metrics := ap.GetMetrics()
	return metrics.ErrorRate < 50.0
}
