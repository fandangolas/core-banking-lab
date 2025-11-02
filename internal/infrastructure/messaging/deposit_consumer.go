package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"bank-api/internal/infrastructure/database"
	"bank-api/internal/infrastructure/database/postgres"
	"bank-api/internal/infrastructure/messaging/kafka"
	"bank-api/internal/pkg/logging"
	"bank-api/internal/pkg/telemetry"

	"github.com/IBM/sarama"
)

// DepositConsumer processes deposit request events from Kafka
type DepositConsumer struct {
	consumerGroup sarama.ConsumerGroup
	publisher     EventPublisher
	db            database.Repository
	config        *kafka.Config
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewDepositConsumer creates a new deposit consumer
func NewDepositConsumer(config *kafka.Config, publisher EventPublisher, db database.Repository) (*DepositConsumer, error) {
	saramaConfig, err := config.ToSaramaConfig()
	if err != nil {
		return nil, err
	}

	// Consumer-specific configuration for at-least-once delivery
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Return.Errors = true

	// At-least-once: Disable auto-commit, commit manually after successful processing
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = false

	// At-least-once: Always read committed messages from the beginning
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}

	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, "deposit-processor-group", saramaConfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &DepositConsumer{
		consumerGroup: consumerGroup,
		publisher:     publisher,
		db:            db,
		config:        config,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Start begins consuming deposit request events
func (c *DepositConsumer) Start() error {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		handler := &depositConsumerHandler{
			publisher: c.publisher,
			db:        c.db,
		}

		topics := []string{kafka.TopicDepositRequests}

		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.consumerGroup.Consume(c.ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			// check if context was cancelled, signaling that the consumer should stop
			if c.ctx.Err() != nil {
				return
			}
		}
	}()

	// Handle errors in a separate goroutine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case err, ok := <-c.consumerGroup.Errors():
				if !ok {
					return
				}
				log.Printf("Consumer group error: %v", err)
			case <-c.ctx.Done():
				return
			}
		}
	}()

	log.Printf("Deposit consumer started: group=deposit-processor-group, topic=%s", kafka.TopicDepositRequests)
	return nil
}

// Stop gracefully stops the consumer
func (c *DepositConsumer) Stop() error {
	c.cancel()
	c.wg.Wait()

	if err := c.consumerGroup.Close(); err != nil {
		return err
	}

	log.Println("Deposit consumer stopped")
	return nil
}

// depositConsumerHandler implements sarama.ConsumerGroupHandler
type depositConsumerHandler struct {
	publisher EventPublisher
	db        database.Repository
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *depositConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *depositConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *depositConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Process the deposit request
			if err := h.processDepositRequest(message); err != nil {
				log.Printf("Failed to process deposit request: offset=%d, error=%v", message.Offset, err)
				// AT-LEAST-ONCE: Don't mark or commit on failure
				// Message will be reprocessed after consumer restart/rebalance
				continue
			}

			// AT-LEAST-ONCE: Mark message and commit immediately after successful processing
			// This ensures we don't reprocess successfully handled messages
			session.MarkMessage(message, "")
			session.Commit() // Explicit commit for at-least-once guarantee

		case <-session.Context().Done():
			return nil
		}
	}
}

// processDepositRequest processes a single deposit request event with idempotency
func (h *depositConsumerHandler) processDepositRequest(message *sarama.ConsumerMessage) error {
	// Deserialize the event
	var event DepositRequestedEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		logging.Error("Failed to unmarshal deposit request event", err, map[string]interface{}{
			"offset": message.Offset,
		})
		return err
	}

	log.Printf("Processing deposit request: operation_id=%s, idempotency_key=%s, account_id=%d, amount=%d",
		event.OperationID, event.IdempotencyKey, event.AccountID, event.Amount)

	// Perform atomic deposit with idempotency check
	// This is THE KEY OPERATION that makes the consumer idempotent!
	acc, err := h.db.AtomicDepositWithIdempotency(event.AccountID, event.Amount, event.IdempotencyKey)

	if err != nil {
		// Check if this is a duplicate operation (expected with at-least-once)
		if errors.Is(err, postgres.ErrDuplicateOperation) {
			log.Printf("Duplicate operation detected (idempotent): idempotency_key=%s, account_id=%d - skipping",
				event.IdempotencyKey, event.AccountID)
			metrics.RecordBankingOperation("deposit", "duplicate")
			return nil // Success! This is idempotent behavior
		}

		// Check if account doesn't exist
		if errors.Is(err, postgres.ErrAccountNotFound) {
			// Publish transaction failed event
			failedEvent := TransactionFailedEvent{
				TransactionType: "deposit",
				AccountID:       event.AccountID,
				Amount:          event.Amount,
				ErrorMessage:    "Account not found",
				Timestamp:       time.Now(),
			}
			if err := h.publisher.PublishTransactionFailed(failedEvent); err != nil {
				logging.Error("Failed to publish transaction failed event", err, map[string]interface{}{
					"operation_id": event.OperationID,
				})
			}
			metrics.RecordBankingOperation("deposit", "error")
			return nil // Don't retry - account doesn't exist
		}

		// Real error - log and retry
		logging.Error("Failed to process deposit", err, map[string]interface{}{
			"operation_id":    event.OperationID,
			"idempotency_key": event.IdempotencyKey,
			"account_id":      event.AccountID,
		})
		metrics.RecordBankingOperation("deposit", "error")
		return err // Retry on database failure
	}

	// Success! Deposit processed atomically
	balance := acc.Balance

	// Record successful operation and metrics
	metrics.RecordBankingOperation("deposit", "success")
	metrics.RecordAccountBalance(float64(balance))

	// Publish deposit completed event
	completedEvent := DepositCompletedEvent{
		AccountID:    event.AccountID,
		Amount:       event.Amount,
		BalanceAfter: balance,
		Timestamp:    time.Now(),
	}
	if err := h.publisher.PublishDepositCompleted(completedEvent); err != nil {
		logging.Error("Failed to publish deposit completed event", err, map[string]interface{}{
			"operation_id": event.OperationID,
			"account_id":   event.AccountID,
		})
		return err // Retry on publish failure
	}

	log.Printf("Deposit processed successfully: operation_id=%s, idempotency_key=%s, account_id=%d, new_balance=%d",
		event.OperationID, event.IdempotencyKey, event.AccountID, balance)

	return nil
}
