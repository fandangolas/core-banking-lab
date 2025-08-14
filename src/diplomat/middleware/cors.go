package middleware

import (
	"bank-api/src/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS adds Cross-Origin Resource Sharing headers to each response
// allowing the dashboard to communicate with the API from configured origins.
func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range cfg.CORS.AllowOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				break
			}
		}

		if !allowed && len(cfg.CORS.AllowOrigins) > 0 {
			// If origin not allowed, set to first allowed origin (fallback)
			c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.CORS.AllowOrigins[0])
		}

		if cfg.CORS.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			strings.Join(cfg.CORS.AllowHeaders, ", "),
		)
		c.Writer.Header().Set(
			"Access-Control-Allow-Methods",
			strings.Join(cfg.CORS.AllowMethods, ", "),
		)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
