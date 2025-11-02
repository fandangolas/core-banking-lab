package messaging

import "sync"

// EventCapture is an in-memory event publisher for testing
// It captures all published events and allows verification in tests
type EventCapture struct {
	accountCreated      []AccountCreatedEvent
	depositRequested    []DepositRequestedEvent
	depositCompleted    []DepositCompletedEvent
	withdrawalCompleted []WithdrawalCompletedEvent
	transferCompleted   []TransferCompletedEvent
	transactionFailed   []TransactionFailedEvent
	mu                  sync.RWMutex
}

// NewEventCapture creates a new event capture publisher
func NewEventCapture() *EventCapture {
	return &EventCapture{
		accountCreated:      make([]AccountCreatedEvent, 0),
		depositRequested:    make([]DepositRequestedEvent, 0),
		depositCompleted:    make([]DepositCompletedEvent, 0),
		withdrawalCompleted: make([]WithdrawalCompletedEvent, 0),
		transferCompleted:   make([]TransferCompletedEvent, 0),
		transactionFailed:   make([]TransactionFailedEvent, 0),
	}
}

// PublishAccountCreated captures account created event
func (e *EventCapture) PublishAccountCreated(event AccountCreatedEvent) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.accountCreated = append(e.accountCreated, event)
	return nil
}

// PublishDepositRequested captures deposit requested event
func (e *EventCapture) PublishDepositRequested(event DepositRequestedEvent) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.depositRequested = append(e.depositRequested, event)
	return nil
}

// PublishDepositCompleted captures deposit completed event
func (e *EventCapture) PublishDepositCompleted(event DepositCompletedEvent) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.depositCompleted = append(e.depositCompleted, event)
	return nil
}

// PublishWithdrawalCompleted captures withdrawal completed event
func (e *EventCapture) PublishWithdrawalCompleted(event WithdrawalCompletedEvent) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.withdrawalCompleted = append(e.withdrawalCompleted, event)
	return nil
}

// PublishTransferCompleted captures transfer completed event
func (e *EventCapture) PublishTransferCompleted(event TransferCompletedEvent) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.transferCompleted = append(e.transferCompleted, event)
	return nil
}

// PublishTransactionFailed captures transaction failed event
func (e *EventCapture) PublishTransactionFailed(event TransactionFailedEvent) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.transactionFailed = append(e.transactionFailed, event)
	return nil
}

// Close is a no-op for event capture
func (e *EventCapture) Close() error {
	return nil
}

// IsHealthy always returns true for event capture
func (e *EventCapture) IsHealthy() bool {
	return true
}

// GetAccountCreatedEvents returns all captured account created events
func (e *EventCapture) GetAccountCreatedEvents() []AccountCreatedEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// Return a copy to prevent modification
	events := make([]AccountCreatedEvent, len(e.accountCreated))
	copy(events, e.accountCreated)
	return events
}

// GetDepositRequestedEvents returns all captured deposit requested events
func (e *EventCapture) GetDepositRequestedEvents() []DepositRequestedEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	events := make([]DepositRequestedEvent, len(e.depositRequested))
	copy(events, e.depositRequested)
	return events
}

// GetDepositCompletedEvents returns all captured deposit completed events
func (e *EventCapture) GetDepositCompletedEvents() []DepositCompletedEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	events := make([]DepositCompletedEvent, len(e.depositCompleted))
	copy(events, e.depositCompleted)
	return events
}

// GetWithdrawalCompletedEvents returns all captured withdrawal completed events
func (e *EventCapture) GetWithdrawalCompletedEvents() []WithdrawalCompletedEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	events := make([]WithdrawalCompletedEvent, len(e.withdrawalCompleted))
	copy(events, e.withdrawalCompleted)
	return events
}

// GetTransferCompletedEvents returns all captured transfer completed events
func (e *EventCapture) GetTransferCompletedEvents() []TransferCompletedEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	events := make([]TransferCompletedEvent, len(e.transferCompleted))
	copy(events, e.transferCompleted)
	return events
}

// GetTransactionFailedEvents returns all captured transaction failed events
func (e *EventCapture) GetTransactionFailedEvents() []TransactionFailedEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	events := make([]TransactionFailedEvent, len(e.transactionFailed))
	copy(events, e.transactionFailed)
	return events
}

// Reset clears all captured events (useful between tests)
func (e *EventCapture) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.accountCreated = make([]AccountCreatedEvent, 0)
	e.depositRequested = make([]DepositRequestedEvent, 0)
	e.depositCompleted = make([]DepositCompletedEvent, 0)
	e.withdrawalCompleted = make([]WithdrawalCompletedEvent, 0)
	e.transferCompleted = make([]TransferCompletedEvent, 0)
	e.transactionFailed = make([]TransactionFailedEvent, 0)
}

// GetEventCount returns the total number of events captured
func (e *EventCapture) GetEventCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.accountCreated) + len(e.depositRequested) +
		len(e.depositCompleted) + len(e.withdrawalCompleted) +
		len(e.transferCompleted) + len(e.transactionFailed)
}
