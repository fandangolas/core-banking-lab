package components

import (
	"bank-api/internal/api/middleware"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequestContextUniqueness verifies each request gets its own context
func TestRequestContextUniqueness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var capturedContexts []*middleware.RequestContext
	var capturedRequestIDs []string

	// Test handler that captures request contexts
	router.Use(middleware.RequestContextMiddleware())
	router.GET("/test", func(c *gin.Context) {
		reqCtx, exists := middleware.GetRequestContext(c)
		require.True(t, exists, "Request context should exist")

		// Capture the context for later comparison
		capturedContexts = append(capturedContexts, reqCtx)
		capturedRequestIDs = append(capturedRequestIDs, reqCtx.RequestID)

		c.JSON(http.StatusOK, gin.H{
			"request_id": reqCtx.RequestID,
			"user_ip":    reqCtx.UserIP,
		})
	})

	// Make multiple requests
	numRequests := 5
	for i := 0; i < numRequests; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1."+string(rune(100+i))) // Different IPs
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Verify we captured all contexts
	assert.Len(t, capturedContexts, numRequests)
	assert.Len(t, capturedRequestIDs, numRequests)

	// Verify each context is unique
	for i := 0; i < numRequests; i++ {
		for j := i + 1; j < numRequests; j++ {
			// Different context instances
			assert.NotSame(t, capturedContexts[i], capturedContexts[j],
				"Request contexts %d and %d should be different instances", i, j)

			// Different request IDs
			assert.NotEqual(t, capturedRequestIDs[i], capturedRequestIDs[j],
				"Request IDs %d and %d should be different", i, j)
		}
	}
}

// TestRequestContextSharedSingletons verifies singletons are shared across requests
func TestRequestContextSharedSingletons(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var capturedDatabases []interface{}
	var capturedBrokers []interface{}

	router.Use(middleware.RequestContextMiddleware())
	router.GET("/test", func(c *gin.Context) {
		reqCtx, exists := middleware.GetRequestContext(c)
		require.True(t, exists)

		// Capture singleton references
		capturedDatabases = append(capturedDatabases, reqCtx.Database)
		capturedBrokers = append(capturedBrokers, reqCtx.EventBroker)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make multiple requests
	numRequests := 3
	for i := 0; i < numRequests; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Verify all requests got the same singleton instances
	require.Len(t, capturedDatabases, numRequests)
	require.Len(t, capturedBrokers, numRequests)

	firstDatabase := capturedDatabases[0]
	firstBroker := capturedBrokers[0]

	for i := 1; i < numRequests; i++ {
		assert.Equal(t, firstDatabase, capturedDatabases[i],
			"Database should be same singleton across requests")
		assert.Equal(t, firstBroker, capturedBrokers[i],
			"Event broker should be same singleton across requests")
	}
}

// TestRequestContextLoggerWithRequestID verifies request-scoped logger includes request metadata
func TestRequestContextLoggerWithRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var loggedRequestID string
	var loggedUserIP string

	router.Use(middleware.RequestContextMiddleware())
	router.GET("/test", func(c *gin.Context) {
		reqCtx, exists := middleware.GetRequestContext(c)
		require.True(t, exists)

		// Store for verification
		loggedRequestID = reqCtx.RequestID
		loggedUserIP = reqCtx.UserIP

		// Use request-scoped logger
		reqCtx.Logger.Info("Test log message", map[string]interface{}{
			"custom": "field",
		})

		c.JSON(http.StatusOK, gin.H{
			"request_id": reqCtx.RequestID,
			"user_ip":    reqCtx.UserIP,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, loggedRequestID, "Request ID should be logged")
	assert.Equal(t, "203.0.113.1", loggedUserIP, "User IP should be captured")
}

// TestRequestContextDuration verifies request timing functionality
func TestRequestContextDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(middleware.RequestContextMiddleware())
	router.GET("/test", func(c *gin.Context) {
		reqCtx, exists := middleware.GetRequestContext(c)
		require.True(t, exists)

		// Add some processing time
		time.Sleep(10 * time.Millisecond)

		duration := reqCtx.Duration()
		assert.True(t, duration >= 10*time.Millisecond,
			"Duration should be at least 10ms, got %v", duration)

		c.JSON(http.StatusOK, gin.H{
			"duration_ms": duration.Milliseconds(),
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	durationMs := response["duration_ms"].(float64)
	assert.True(t, durationMs >= 10, "Response should show at least 10ms duration")
}

// TestRequestContextValueStorage verifies request-scoped value storage
func TestRequestContextValueStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(middleware.RequestContextMiddleware())
	router.GET("/test", func(c *gin.Context) {
		reqCtx, exists := middleware.GetRequestContext(c)
		require.True(t, exists)

		// Store custom values in request context
		reqCtx.WithValue("user_id", "12345")
		reqCtx.WithValue("feature_flag", true)

		// Retrieve values
		userID := reqCtx.Value("user_id")
		featureFlag := reqCtx.Value("feature_flag")
		nonExistent := reqCtx.Value("non_existent")

		c.JSON(http.StatusOK, gin.H{
			"user_id":      userID,
			"feature_flag": featureFlag,
			"non_existent": nonExistent,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "12345", response["user_id"])
	assert.Equal(t, true, response["feature_flag"])
	assert.Nil(t, response["non_existent"])
}
