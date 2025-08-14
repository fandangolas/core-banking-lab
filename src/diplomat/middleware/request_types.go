package middleware

import (
	"bank-api/src/diplomat/database"
	"bank-api/src/diplomat/events"
	"bank-api/src/logging"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestContext holds request-scoped dependencies and context
// This is created fresh for each HTTP request
type RequestContext struct {
	// Request metadata
	RequestID   string
	UserIP      string
	UserAgent   string
	StartTime   time.Time
	GinContext  *gin.Context
	Context     context.Context
	
	// Request-scoped services (these reference the singletons)
	Database    database.Repository
	EventBroker *events.Broker
	Logger      RequestLogger
}

// RequestLogger provides request-scoped logging with automatic field injection
type RequestLogger struct {
	requestID string
	userIP    string
}

// NewRequestContext creates a new request-scoped context
// This should be called at the beginning of each HTTP handler
func NewRequestContext(ginCtx *gin.Context) *RequestContext {
	requestID := uuid.New().String()
	
	// Create request context with timeout
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	
	return &RequestContext{
		RequestID:   requestID,
		UserIP:      ginCtx.ClientIP(),
		UserAgent:   ginCtx.GetHeader("User-Agent"),
		StartTime:   time.Now(),
		GinContext:  ginCtx,
		Context:     ctx,
		
		// Reference the singleton services
		Database:    database.Repo,
		EventBroker: events.GetBroker(),
		Logger: RequestLogger{
			requestID: requestID,
			userIP:    ginCtx.ClientIP(),
		},
	}
}

// Info logs info level with request context automatically injected
func (rl RequestLogger) Info(message string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["request_id"] = rl.requestID
	fields["user_ip"] = rl.userIP
	
	logging.Info(message, fields)
}

// Warn logs warning level with request context automatically injected
func (rl RequestLogger) Warn(message string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["request_id"] = rl.requestID
	fields["user_ip"] = rl.userIP
	
	logging.Warn(message, fields)
}

// Error logs error level with request context automatically injected
func (rl RequestLogger) Error(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["request_id"] = rl.requestID
	fields["user_ip"] = rl.userIP
	
	logging.Error(message, err, fields)
}

// Duration returns how long this request has been processing
func (rc *RequestContext) Duration() time.Duration {
	return time.Since(rc.StartTime)
}

// WithValue adds a value to the request context
func (rc *RequestContext) WithValue(key, value interface{}) {
	rc.Context = context.WithValue(rc.Context, key, value)
}

// Value retrieves a value from the request context  
func (rc *RequestContext) Value(key interface{}) interface{} {
	return rc.Context.Value(key)
}

// Finish should be called at the end of request processing for cleanup/metrics
func (rc *RequestContext) Finish() {
	duration := rc.Duration()
	rc.Logger.Info("Request completed", map[string]interface{}{
		"duration_ms": duration.Milliseconds(),
		"method":      rc.GinContext.Request.Method,
		"path":        rc.GinContext.Request.URL.Path,
		"status":      rc.GinContext.Writer.Status(),
	})
}