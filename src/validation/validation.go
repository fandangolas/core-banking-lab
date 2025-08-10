package validation

import (
	"errors"
	"strings"
	"unicode"
)

const (
	MinAmount     = 1
	MaxAmount     = 1000000  // R$ 10,000.00 (in centavos)
	MaxOwnerLen   = 100
	MinOwnerLen   = 2
)

func ValidateAmount(amount int) error {
	if amount < MinAmount {
		return errors.New("amount must be greater than zero")
	}
	if amount > MaxAmount {
		return errors.New("amount exceeds maximum limit of R$ 10,000.00")
	}
	return nil
}

func ValidateOwnerName(owner string) error {
	owner = strings.TrimSpace(owner)
	
	if len(owner) < MinOwnerLen {
		return errors.New("owner name must be at least 2 characters")
	}
	
	if len(owner) > MaxOwnerLen {
		return errors.New("owner name cannot exceed 100 characters")
	}
	
	// Check if name contains only letters, spaces, and common punctuation
	for _, r := range owner {
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) && r != '.' && r != '-' && r != '\'' {
			return errors.New("owner name contains invalid characters")
		}
	}
	
	return nil
}

func ValidateAccountID(id int) error {
	if id <= 0 {
		return errors.New("account ID must be positive")
	}
	return nil
}