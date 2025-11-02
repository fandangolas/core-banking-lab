package messaging

import "time"

// AccountCreatedEvent represents an account creation event
type AccountCreatedEvent struct {
	AccountID int       `json:"account_id"`
	Owner     string    `json:"owner"`
	Timestamp time.Time `json:"timestamp"`
}

// DepositRequestedEvent represents a deposit command request
type DepositRequestedEvent struct {
	OperationID    string    `json:"operation_id"`    // UUID for tracking (legacy)
	IdempotencyKey string    `json:"idempotency_key"` // SHA-256 hash for deduplication
	AccountID      int       `json:"account_id"`
	Amount         int       `json:"amount"` // in cents
	Timestamp      time.Time `json:"timestamp"`
}

// DepositCompletedEvent represents a successful deposit
type DepositCompletedEvent struct {
	AccountID    int       `json:"account_id"`
	Amount       int       `json:"amount"`        // in cents
	BalanceAfter int       `json:"balance_after"` // in cents
	Timestamp    time.Time `json:"timestamp"`
}

// WithdrawalCompletedEvent represents a successful withdrawal
type WithdrawalCompletedEvent struct {
	AccountID    int       `json:"account_id"`
	Amount       int       `json:"amount"`        // in cents
	BalanceAfter int       `json:"balance_after"` // in cents
	Timestamp    time.Time `json:"timestamp"`
}

// TransferCompletedEvent represents a successful transfer
type TransferCompletedEvent struct {
	FromAccountID    int       `json:"from_account_id"`
	ToAccountID      int       `json:"to_account_id"`
	Amount           int       `json:"amount"`             // in cents
	FromBalanceAfter int       `json:"from_balance_after"` // in cents
	ToBalanceAfter   int       `json:"to_balance_after"`   // in cents
	Timestamp        time.Time `json:"timestamp"`
}

// TransactionFailedEvent represents a failed transaction for audit trail
type TransactionFailedEvent struct {
	TransactionType string    `json:"transaction_type"` // deposit, withdrawal, transfer
	AccountID       int       `json:"account_id,omitempty"`
	FromAccountID   int       `json:"from_account_id,omitempty"`
	ToAccountID     int       `json:"to_account_id,omitempty"`
	Amount          int       `json:"amount"` // in cents
	ErrorMessage    string    `json:"error_message"`
	Timestamp       time.Time `json:"timestamp"`
}
