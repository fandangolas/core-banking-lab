package handlers

import (
	"bank-api/internal/infrastructure/database"
	"bank-api/internal/infrastructure/messaging"
)

// HandlerDependencies is an interface that defines the dependencies needed by handlers
// This interface breaks the circular dependency between handlers and components packages
type HandlerDependencies interface {
	GetDatabase() database.Repository
	GetEventPublisher() messaging.EventPublisher
}
