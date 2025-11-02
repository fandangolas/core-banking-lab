package middleware

import (
	"bank-api/internal/api/handlers"
	"bank-api/internal/infrastructure/messaging"

	"github.com/gin-gonic/gin"
)

// EventPublisherMiddleware injects the event publisher into the request context
func EventPublisherMiddleware(publisher messaging.EventPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		handlers.SetEventPublisher(c, publisher)
		c.Next()
	}
}
