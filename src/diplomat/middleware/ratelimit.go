package middleware

import (
	"bank-api/src/config"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func RateLimit(cfg *config.Config) gin.HandlerFunc {
	limiter := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    cfg.RateLimit.RequestsPerMinute,
		window:   cfg.RateLimit.Window,
	}
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		limiter.mutex.Lock()
		defer limiter.mutex.Unlock()

		now := time.Now()
		
		// Clean old requests outside the window
		if requests, exists := limiter.requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < limiter.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			limiter.requests[clientIP] = validRequests
		}

		// Check if limit exceeded
		if len(limiter.requests[clientIP]) >= limiter.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Try again later.",
				"retry_after": int(limiter.window.Seconds()),
			})
			c.Abort()
			return
		}

		// Add current request
		limiter.requests[clientIP] = append(limiter.requests[clientIP], now)
		
		c.Next()
	}
}