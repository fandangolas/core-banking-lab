package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

// Producer wraps Kafka producer for event publishing
type Producer struct {
	producer sarama.SyncProducer
	config   *Config
	mu       sync.RWMutex
	closed   bool
}

// NewProducer creates a new Kafka producer
func NewProducer(config *Config) (*Producer, error) {
	saramaConfig, err := config.ToSaramaConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama config: %w", err)
	}

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	log.Printf("Kafka producer initialized: brokers=%v, client_id=%s", config.Brokers, config.ClientID)

	return &Producer{
		producer: producer,
		config:   config,
	}, nil
}

// PublishEvent publishes an event to a Kafka topic
func (p *Producer) PublishEvent(topic string, key string, event interface{}) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer is closed")
	}
	p.mu.RUnlock()

	// Serialize event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(eventJSON),
	}

	// Send message (synchronous)
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to publish event to Kafka: topic=%s, key=%s, error=%v", topic, key, err)
		return fmt.Errorf("failed to send message to kafka: %w", err)
	}

	log.Printf("Event published to Kafka: topic=%s, partition=%d, offset=%d, key=%s", topic, partition, offset, key)
	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("failed to close kafka producer: %w", err)
	}

	log.Println("Kafka producer closed")
	return nil
}

// IsHealthy checks if the producer is healthy
func (p *Producer) IsHealthy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return !p.closed
}
