package handlers

import (
	"bank-api/src/diplomat/events"
	"io"

	"github.com/gin-gonic/gin"
)

func Events(c *gin.Context) {
	ch := events.BrokerInstance.Subscribe()
	defer events.BrokerInstance.Unsubscribe(ch)

	c.Stream(func(w io.Writer) bool {
		if evt, ok := <-ch; ok {
			c.SSEvent("transaction", evt)
			return true
		}
		return false
	})
}
