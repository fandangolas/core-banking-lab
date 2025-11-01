package handlers

import (
	"bank-api/internal/infrastructure/events"
	"io"

	"github.com/gin-gonic/gin"
)

func Events(c *gin.Context) {
	broker := events.GetBroker()
	ch := broker.Subscribe()
	defer broker.Unsubscribe(ch)

	c.Stream(func(w io.Writer) bool {
		if evt, ok := <-ch; ok {
			c.SSEvent("transaction", evt)
			return true
		}
		return false
	})
}
