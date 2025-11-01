package models

import "time"

type TransactionEvent struct {
	Type        string    `json:"type"`
	AccountID   int       `json:"account_id,omitempty"`
	FromID      int       `json:"from_id,omitempty"`
	ToID        int       `json:"to_id,omitempty"`
	Amount      int       `json:"amount"`
	Balance     int       `json:"balance,omitempty"`
	FromBalance int       `json:"from_balance,omitempty"`
	ToBalance   int       `json:"to_balance,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}
