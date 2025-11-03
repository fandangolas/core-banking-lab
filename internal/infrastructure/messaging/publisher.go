package messaging

import (
	"fmt"
	"strconv"

	"bank-api/internal/infrastructure/messaging/kafka"
)

// EventPublisher defines the interface for publishing banking events
type EventPublisher interface {
	PublishAccountCreated(event AccountCreatedEvent) error
	PublishDepositRequested(event DepositRequestedEvent) error
	PublishDepositCompleted(event DepositCompletedEvent) error
	PublishWithdrawalCompleted(event WithdrawalCompletedEvent) error
	PublishTransferCompleted(event TransferCompletedEvent) error
	PublishTransactionFailed(event TransactionFailedEvent) error
	Close() error
	IsHealthy() bool
}

// KafkaEventPublisher implements EventPublisher using Kafka with async producer
type KafkaEventPublisher struct {
	producer *kafka.AsyncProducer
}

// NewKafkaEventPublisher creates a new high-performance async Kafka event publisher
func NewKafkaEventPublisher(config *kafka.Config) (*KafkaEventPublisher, error) {
	producer, err := kafka.NewAsyncProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka async producer: %w", err)
	}

	return &KafkaEventPublisher{
		producer: producer,
	}, nil
}

// PublishAccountCreated publishes an account created event (async)
func (p *KafkaEventPublisher) PublishAccountCreated(event AccountCreatedEvent) error {
	key := strconv.Itoa(event.AccountID)
	return p.producer.PublishEventAsync(kafka.TopicAccountCreated, key, event)
}

// PublishDepositRequested publishes a deposit request command (async)
func (p *KafkaEventPublisher) PublishDepositRequested(event DepositRequestedEvent) error {
	key := strconv.Itoa(event.AccountID)
	return p.producer.PublishEventAsync(kafka.TopicDepositRequests, key, event)
}

// PublishDepositCompleted publishes a deposit completed event (async)
func (p *KafkaEventPublisher) PublishDepositCompleted(event DepositCompletedEvent) error {
	key := strconv.Itoa(event.AccountID)
	return p.producer.PublishEventAsync(kafka.TopicTransactionDeposit, key, event)
}

// PublishWithdrawalCompleted publishes a withdrawal completed event (async)
func (p *KafkaEventPublisher) PublishWithdrawalCompleted(event WithdrawalCompletedEvent) error {
	key := strconv.Itoa(event.AccountID)
	return p.producer.PublishEventAsync(kafka.TopicTransactionWithdrawal, key, event)
}

// PublishTransferCompleted publishes a transfer completed event (async)
func (p *KafkaEventPublisher) PublishTransferCompleted(event TransferCompletedEvent) error {
	key := fmt.Sprintf("%d-%d", event.FromAccountID, event.ToAccountID)
	return p.producer.PublishEventAsync(kafka.TopicTransactionTransfer, key, event)
}

// PublishTransactionFailed publishes a transaction failed event (async)
func (p *KafkaEventPublisher) PublishTransactionFailed(event TransactionFailedEvent) error {
	// Use account ID as key if available, otherwise use transaction type
	key := event.TransactionType
	if event.AccountID != 0 {
		key = strconv.Itoa(event.AccountID)
	} else if event.FromAccountID != 0 {
		key = strconv.Itoa(event.FromAccountID)
	}
	return p.producer.PublishEventAsync(kafka.TopicTransactionFailed, key, event)
}

// Close closes the Kafka producer
func (p *KafkaEventPublisher) Close() error {
	return p.producer.Close()
}

// IsHealthy checks if the publisher is healthy
func (p *KafkaEventPublisher) IsHealthy() bool {
	return p.producer.IsHealthy()
}

// NoOpEventPublisher is a no-op implementation for testing
type NoOpEventPublisher struct{}

// NewNoOpEventPublisher creates a no-op event publisher
func NewNoOpEventPublisher() *NoOpEventPublisher {
	return &NoOpEventPublisher{}
}

func (p *NoOpEventPublisher) PublishAccountCreated(event AccountCreatedEvent) error     { return nil }
func (p *NoOpEventPublisher) PublishDepositRequested(event DepositRequestedEvent) error { return nil }
func (p *NoOpEventPublisher) PublishDepositCompleted(event DepositCompletedEvent) error { return nil }
func (p *NoOpEventPublisher) PublishWithdrawalCompleted(event WithdrawalCompletedEvent) error {
	return nil
}
func (p *NoOpEventPublisher) PublishTransferCompleted(event TransferCompletedEvent) error { return nil }
func (p *NoOpEventPublisher) PublishTransactionFailed(event TransactionFailedEvent) error { return nil }
func (p *NoOpEventPublisher) Close() error                                                { return nil }
func (p *NoOpEventPublisher) IsHealthy() bool                                             { return true }
