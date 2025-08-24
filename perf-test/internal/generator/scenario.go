package generator

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type OperationType string

const (
	OpDeposit  OperationType = "deposit"
	OpWithdraw OperationType = "withdraw"
	OpTransfer OperationType = "transfer"
	OpBalance  OperationType = "balance"
)

type Scenario struct {
	Name             string                    `json:"name"`
	Description      string                    `json:"description"`
	Accounts         int                       `json:"accounts"`
	TargetOperations int64                     `json:"target_operations"`
	Operations       []Operation               `json:"operations"`
	Distribution     map[OperationType]float64 `json:"distribution"`
	InitialBalance   float64                   `json:"initial_balance"`
	MinAmount        float64                   `json:"min_amount"`
	MaxAmount        float64                   `json:"max_amount"`
	ThinkTime        time.Duration             `json:"think_time"`
}

type Operation struct {
	Type      OperationType `json:"type"`
	AccountID string        `json:"account_id,omitempty"`
	FromID    string        `json:"from_id,omitempty"`
	ToID      string        `json:"to_id,omitempty"`
	Amount    float64       `json:"amount,omitempty"`
}

func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	var scenario Scenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse scenario: %w", err)
	}

	if err := scenario.Validate(); err != nil {
		return nil, fmt.Errorf("invalid scenario: %w", err)
	}

	return &scenario, nil
}

func (s *Scenario) Validate() error {
	if s.Accounts <= 0 {
		return fmt.Errorf("accounts must be positive")
	}

	total := 0.0
	for _, weight := range s.Distribution {
		total += weight
	}

	if total < 0.99 || total > 1.01 {
		return fmt.Errorf("distribution weights must sum to 1.0")
	}

	return nil
}

func (s *Scenario) GenerateOperation(accountIDs []string) Operation {
	r := rand.Float64()
	cumulative := 0.0

	for opType, weight := range s.Distribution {
		cumulative += weight
		if r <= cumulative {
			return s.createOperation(opType, accountIDs)
		}
	}

	return s.createOperation(OpBalance, accountIDs)
}

func (s *Scenario) createOperation(opType OperationType, accountIDs []string) Operation {
	op := Operation{Type: opType}

	switch opType {
	case OpDeposit, OpWithdraw:
		op.AccountID = accountIDs[rand.Intn(len(accountIDs))]
		op.Amount = s.generateValidAmount()
	case OpTransfer:
		fromIdx := rand.Intn(len(accountIDs))
		toIdx := rand.Intn(len(accountIDs))
		for toIdx == fromIdx && len(accountIDs) > 1 {
			toIdx = rand.Intn(len(accountIDs))
		}
		op.FromID = accountIDs[fromIdx]
		op.ToID = accountIDs[toIdx]
		op.Amount = s.generateValidAmount()
	case OpBalance:
		op.AccountID = accountIDs[rand.Intn(len(accountIDs))]
	}

	return op
}

func (s *Scenario) generateValidAmount() float64 {
	// Generate amount in cents between MinAmount*100 and MaxAmount*100
	minCents := int(s.MinAmount * 100)
	maxCents := int(s.MaxAmount * 100)
	
	// Ensure minimum of 1 cent
	if minCents < 1 {
		minCents = 1
	}
	
	// Generate random amount in cents
	cents := minCents + rand.Intn(maxCents-minCents+1)
	
	// Convert back to float (dollars) for display, but executor will convert to int
	return float64(cents)
}

func DefaultScenario() *Scenario {
	return &Scenario{
		Name:        "Default Banking Load Test",
		Description: "Balanced mix of banking operations with realistic amounts",
		Accounts:    1000,
		Distribution: map[OperationType]float64{
			OpDeposit:  0.25,
			OpWithdraw: 0.25,
			OpTransfer: 0.35,
			OpBalance:  0.15,
		},
		InitialBalance: 100000.00, // 1000.00 in dollars (100000 cents)
		MinAmount:      1.00,      // 1.00 in dollars (100 cents)
		MaxAmount:      10.00,     // 10.00 in dollars (1000 cents)
		ThinkTime:      10 * time.Millisecond,
	}
}

func HighConcurrencyScenario() *Scenario {
	return &Scenario{
		Name:        "High Concurrency Transfer Test",
		Description: "Heavy transfer load to test deadlock prevention",
		Accounts:    100,
		Distribution: map[OperationType]float64{
			OpDeposit:  0.10,
			OpWithdraw: 0.10,
			OpTransfer: 0.70,
			OpBalance:  0.10,
		},
		InitialBalance: 50000.00,
		MinAmount:      100.00,
		MaxAmount:      5000.00,
		ThinkTime:      1 * time.Millisecond,
	}
}

func ReadHeavyScenario() *Scenario {
	return &Scenario{
		Name:        "Read Heavy Load Test",
		Description: "Mostly balance checks with occasional writes",
		Accounts:    5000,
		Distribution: map[OperationType]float64{
			OpDeposit:  0.05,
			OpWithdraw: 0.05,
			OpTransfer: 0.10,
			OpBalance:  0.80,
		},
		InitialBalance: 1000.00,
		MinAmount:      50.00,
		MaxAmount:      500.00,
		ThinkTime:      5 * time.Millisecond,
	}
}