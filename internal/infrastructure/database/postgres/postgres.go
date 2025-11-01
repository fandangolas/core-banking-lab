package postgres

import (
	"bank-api/internal/domain/models"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements the Repository interface using PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
	mu   sync.RWMutex // Protects account mutex map
	// Account-level mutexes for concurrency control (same as in-memory)
	accountMutexes map[int]*sync.Mutex
}

// NewPostgresRepository creates a new PostgreSQL repository with connection pool
func NewPostgresRepository(cfg *Config) (*PostgresRepository, error) {
	ctx := context.Background()

	// Parse connection string and create pool config
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure connection pool settings from config
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)

	// Parse duration strings
	if maxLifetime, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
		poolConfig.MaxConnLifetime = maxLifetime
	}
	if maxIdleTime, err := time.ParseDuration(cfg.ConnMaxIdleTime); err == nil {
		poolConfig.MaxConnIdleTime = maxIdleTime
	}
	if healthCheck, err := time.ParseDuration(cfg.HealthCheckPeriod); err == nil {
		poolConfig.HealthCheckPeriod = healthCheck
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("PostgreSQL connection pool created successfully (max: %d, min: %d)",
		poolConfig.MaxConns, poolConfig.MinConns)

	return &PostgresRepository{
		pool:           pool,
		accountMutexes: make(map[int]*sync.Mutex),
	}, nil
}

// Close closes the database connection pool
func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
		log.Println("PostgreSQL connection pool closed")
	}
}

// getAccountMutex returns the mutex for a specific account ID
// This maintains the same concurrency control pattern as in-memory implementation
func (r *PostgresRepository) getAccountMutex(accountID int) *sync.Mutex {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accountMutexes[accountID]; !exists {
		r.accountMutexes[accountID] = &sync.Mutex{}
	}
	return r.accountMutexes[accountID]
}

// CreateAccount creates a new account with the given owner
// Returns the ID of the newly created account
func (r *PostgresRepository) CreateAccount(owner string) int {
	ctx := context.Background()

	query := `
		INSERT INTO accounts (owner, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var accountID int
	now := time.Now()

	err := r.pool.QueryRow(ctx, query, owner, 0, now, now).Scan(&accountID)
	if err != nil {
		log.Printf("Failed to create account for owner %s: %v", owner, err)
		return 0
	}

	log.Printf("Account created: ID=%d, Owner=%s", accountID, owner)
	return accountID
}

// GetAccount retrieves an account by ID
// Returns the account and true if found, nil and false otherwise
func (r *PostgresRepository) GetAccount(id int) (*models.Account, bool) {
	ctx := context.Background()

	query := `
		SELECT id, owner, balance, created_at
		FROM accounts
		WHERE id = $1
	`

	var account models.Account
	var balanceDecimal float64

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&account.Id,
		&account.Owner,
		&balanceDecimal,
		&account.CreatedAt,
	)

	if err != nil {
		// Account not found or other error
		return nil, false
	}

	// Convert balance from DECIMAL(15,2) to cents (int)
	account.Balance = int(balanceDecimal * 100)

	return &account, true
}

// UpdateAccount updates an existing account's balance
// This is called after in-memory modifications to persist changes
func (r *PostgresRepository) UpdateAccount(acc *models.Account) {
	ctx := context.Background()

	// Get account-specific mutex to prevent concurrent updates
	mu := r.getAccountMutex(acc.Id)
	mu.Lock()
	defer mu.Unlock()

	query := `
		UPDATE accounts
		SET balance = $1, version = version + 1
		WHERE id = $2
	`

	// Convert balance from cents (int) to DECIMAL(15,2)
	balanceDecimal := float64(acc.Balance) / 100.0

	_, err := r.pool.Exec(ctx, query, balanceDecimal, acc.Id)
	if err != nil {
		log.Printf("Failed to update account %d: %v", acc.Id, err)
		return
	}

	log.Printf("Account updated: ID=%d, Balance=%.2f", acc.Id, balanceDecimal)
}

// Reset clears all data from the database
// WARNING: This is only for testing purposes
func (r *PostgresRepository) Reset() {
	ctx := context.Background()

	// Clear account mutexes
	r.mu.Lock()
	r.accountMutexes = make(map[int]*sync.Mutex)
	r.mu.Unlock()

	// Truncate tables in correct order (transactions first due to foreign key)
	queries := []string{
		"TRUNCATE TABLE transactions RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE accounts RESTART IDENTITY CASCADE",
	}

	for _, query := range queries {
		_, err := r.pool.Exec(ctx, query)
		if err != nil {
			log.Printf("Failed to reset database: %v", err)
			return
		}
	}

	log.Println("Database reset completed")
}

// CreateTransaction records a transaction in the database
// This is called after successful account operations for audit trail
func (r *PostgresRepository) CreateTransaction(accountID int, txType string, amount int, balanceAfter int, referenceID *string) error {
	ctx := context.Background()

	query := `
		INSERT INTO transactions (account_id, transaction_type, amount, balance_after, reference_id)
		VALUES ($1, $2, $3, $4, $5)
	`

	// Convert amounts from cents to DECIMAL(15,2)
	amountDecimal := float64(amount) / 100.0
	balanceAfterDecimal := float64(balanceAfter) / 100.0

	_, err := r.pool.Exec(ctx, query, accountID, txType, amountDecimal, balanceAfterDecimal, referenceID)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetTransactionHistory retrieves the transaction history for an account
// Returns the most recent transactions first
func (r *PostgresRepository) GetTransactionHistory(accountID int, limit int) ([]map[string]interface{}, error) {
	ctx := context.Background()

	query := `
		SELECT id, transaction_type, amount, balance_after, reference_id, created_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []map[string]interface{}

	for rows.Next() {
		var (
			id           int
			txType       string
			amount       float64
			balanceAfter float64
			referenceID  *string
			createdAt    time.Time
		)

		err := rows.Scan(&id, &txType, &amount, &balanceAfter, &referenceID, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		tx := map[string]interface{}{
			"id":            id,
			"type":          txType,
			"amount":        amount,
			"balance_after": balanceAfter,
			"created_at":    createdAt,
		}

		if referenceID != nil {
			tx["reference_id"] = *referenceID
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}
