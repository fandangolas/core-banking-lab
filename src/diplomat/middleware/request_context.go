package middleware

import (
	"bank-api/src/context"

	"github.com/gin-gonic/gin"
)

const RequestContextKey = "request_context"

// RequestContext middleware creates a new request-scoped context for each request
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create request-scoped context
		reqCtx := context.NewRequestContext(c)
		
		// Store in gin context for handlers to access
		c.Set(RequestContextKey, reqCtx)
		
		// Log request start
		reqCtx.Logger.Info("Request started", map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": reqCtx.UserAgent,
		})
		
		// Process request
		c.Next()
		
		// Log request completion and cleanup
		reqCtx.Finish()
	}
}

// GetRequestContext retrieves the request context from gin context
// This helper makes it easy for handlers to access request-scoped services
func GetRequestContext(c *gin.Context) (*context.RequestContext, bool) {
	reqCtx, exists := c.Get(RequestContextKey)
	if !exists {
		return nil, false
	}
	
	requestContext, ok := reqCtx.(*context.RequestContext)
	return requestContext, ok
}