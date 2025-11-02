package handlers

import (
	"bank-api/internal/infrastructure/messaging"

	"github.com/gin-gonic/gin"
)

const eventPublisherKey = "event_publisher"

// SetEventPublisher stores the event publisher in the Gin context
func SetEventPublisher(c *gin.Context, publisher messaging.EventPublisher) {
	c.Set(eventPublisherKey, publisher)
}

// GetEventPublisher retrieves the event publisher from the Gin context
func GetEventPublisher(c *gin.Context) messaging.EventPublisher {
	if publisher, exists := c.Get(eventPublisherKey); exists {
		if ep, ok := publisher.(messaging.EventPublisher); ok {
			return ep
		}
	}
	// Return no-op publisher if not found (safe fallback)
	return messaging.NewNoOpEventPublisher()
}
