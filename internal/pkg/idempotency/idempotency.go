package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateKey creates a deterministic idempotency key from operation details.
// The key is a SHA-256 hash of the operation type, account ID, and amount.
//
// This ensures that:
// - Identical requests produce the same key (consumer can deduplicate)
// - Different requests produce different keys (no false positives)
// - Key generation is fast and doesn't require database access
//
// Examples:
//   - "deposit:1:1000" → "5d41402abc4b2a76b9719d911017c592..."
//   - "deposit:1:1000" → "5d41402abc4b2a76b9719d911017c592..." (same!)
//   - "deposit:1:2000" → "6c8349cc7260ae62e3b1396831a8398f..." (different)
func GenerateKey(operationType string, accountID int, amount int) string {
	// Format: "operation_type:account_id:amount"
	data := fmt.Sprintf("%s:%d:%d", operationType, accountID, amount)

	// SHA-256 hash (collision probability for 1B operations: ~4.3×10^-60)
	hash := sha256.Sum256([]byte(data))

	// Return hex-encoded hash (64 characters)
	return hex.EncodeToString(hash[:])
}

// GenerateTransferKey creates a deterministic idempotency key for transfer operations.
// The key includes both source and destination accounts to ensure uniqueness.
//
// Example:
//   - "transfer:1:2:500" → "a1b2c3d4..." (account 1 → account 2, $5.00)
func GenerateTransferKey(fromAccountID int, toAccountID int, amount int) string {
	// Format: "transfer:from_account:to_account:amount"
	data := fmt.Sprintf("transfer:%d:%d:%d", fromAccountID, toAccountID, amount)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
