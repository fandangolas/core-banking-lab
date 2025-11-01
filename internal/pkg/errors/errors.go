package errors

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e APIError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrCodeValidation        = "VALIDATION_ERROR"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeInternalServer    = "INTERNAL_SERVER_ERROR"
	ErrCodeRateLimit         = "RATE_LIMIT_EXCEEDED"
	ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	ErrCodeInvalidAmount     = "INVALID_AMOUNT"
	ErrCodeAccountNotFound   = "ACCOUNT_NOT_FOUND"
	ErrCodeSelfTransfer      = "SELF_TRANSFER_NOT_ALLOWED"
)

// Error constructors
func NewValidationError(message string) APIError {
	return APIError{
		Code:    ErrCodeValidation,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func NewNotFoundError(resource string) APIError {
	return APIError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Status:  http.StatusNotFound,
	}
}

func NewInternalServerError(message string) APIError {
	return APIError{
		Code:    ErrCodeInternalServer,
		Message: "Internal server error",
		Status:  http.StatusInternalServerError,
	}
}

func NewRateLimitError() APIError {
	return APIError{
		Code:    ErrCodeRateLimit,
		Message: "Rate limit exceeded. Please try again later.",
		Status:  http.StatusTooManyRequests,
	}
}

func NewInsufficientFundsError() APIError {
	return APIError{
		Code:    ErrCodeInsufficientFunds,
		Message: "Insufficient funds for this transaction",
		Status:  http.StatusBadRequest,
	}
}

func NewInvalidAmountError(message string) APIError {
	return APIError{
		Code:    ErrCodeInvalidAmount,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func NewAccountNotFoundError() APIError {
	return APIError{
		Code:    ErrCodeAccountNotFound,
		Message: "Account not found",
		Status:  http.StatusNotFound,
	}
}

func NewSelfTransferError() APIError {
	return APIError{
		Code:    ErrCodeSelfTransfer,
		Message: "Cannot transfer to the same account",
		Status:  http.StatusBadRequest,
	}
}
