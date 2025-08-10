package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS adds Cross-Origin Resource Sharing headers to each response
// allowing the dashboard to communicate with the API from another origin.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Authorization, Accept, X-Requested-With",
		)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, PATCH, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
