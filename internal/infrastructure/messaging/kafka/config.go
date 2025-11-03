package kafka

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

// Config holds Kafka producer configuration
type Config struct {
	Brokers           []string
	ClientID          string
	EnableIdempotence bool
	CompressionType   string
	RequiredAcks      string
	MaxRetries        int
	RetryBackoff      time.Duration
}

// NewConfigFromEnv creates Kafka config from environment variables
func NewConfigFromEnv() *Config {
	brokersStr := getEnv("KAFKA_BROKERS", "localhost:9092")
	brokers := strings.Split(brokersStr, ",")

	return &Config{
		Brokers:           brokers,
		ClientID:          getEnv("KAFKA_CLIENT_ID", "banking-api"),
		EnableIdempotence: getEnvBool("KAFKA_ENABLE_IDEMPOTENCE", false), // Disabled for high throughput - consumer handles idempotency
		CompressionType:   getEnv("KAFKA_COMPRESSION_TYPE", "snappy"),
		RequiredAcks:      getEnv("KAFKA_REQUIRED_ACKS", "all"), // Wait for all in-sync replicas for durability
		MaxRetries:        getEnvInt("KAFKA_MAX_RETRIES", 5),
		RetryBackoff:      getEnvDuration("KAFKA_RETRY_BACKOFF", 100*time.Millisecond),
	}
}

// ToSaramaConfig converts to Sarama configuration
func (c *Config) ToSaramaConfig() (*sarama.Config, error) {
	config := sarama.NewConfig()

	// Producer config
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Idempotent = c.EnableIdempotence
	config.Producer.Retry.Max = c.MaxRetries
	config.Producer.Retry.Backoff = c.RetryBackoff

	// High-throughput producer settings
	// When idempotence is disabled, we can have multiple in-flight requests for better throughput
	if !c.EnableIdempotence {
		config.Net.MaxOpenRequests = 10 // Increase parallelism (was 5)
	} else {
		// Sarama requires MaxOpenRequests=1 when idempotence is enabled
		config.Net.MaxOpenRequests = 1
	}

	// Increase buffer sizes for high-load scenarios
	config.ChannelBufferSize = 100000 // Kafka's internal buffer (was 10,000)

	// Batching configuration for better throughput
	config.Producer.Flush.MaxMessages = 10000     // Larger batches (was 1000)
	config.Producer.Flush.Frequency = 500 * time.Millisecond // More accumulation time (was 100ms)
	config.Producer.Flush.Messages = 1000         // Start flushing after 1000 messages (was 100)

	// Set required acks
	switch c.RequiredAcks {
	case "all", "-1":
		config.Producer.RequiredAcks = sarama.WaitForAll
	case "1":
		config.Producer.RequiredAcks = sarama.WaitForLocal
	case "0":
		config.Producer.RequiredAcks = sarama.NoResponse
	default:
		return nil, fmt.Errorf("invalid required acks value: %s", c.RequiredAcks)
	}

	// Set compression type
	switch c.CompressionType {
	case "none":
		config.Producer.Compression = sarama.CompressionNone
	case "gzip":
		config.Producer.Compression = sarama.CompressionGZIP
	case "snappy":
		config.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		config.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		config.Producer.Compression = sarama.CompressionZSTD
	default:
		return nil, fmt.Errorf("invalid compression type: %s", c.CompressionType)
	}

	// Client ID
	config.ClientID = c.ClientID

	// Version
	config.Version = sarama.V3_0_0_0

	return config, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		fmt.Sscanf(value, "%d", &intValue)
		return intValue
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}
	return defaultValue
}
