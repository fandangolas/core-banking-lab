package routes

import (
	"bank-api/src/diplomat/middleware"
	"bank-api/src/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	router.Use(middleware.RequestContext())      // Add request-scoped context (first!)
	router.Use(middleware.Metrics())
	router.Use(middleware.PrometheusMiddleware()) // Add Prometheus metrics collection

	router.POST("/accounts", handlers.CreateAccount)
	router.GET("/accounts/:id/balance", handlers.GetBalance)
	router.POST("/accounts/:id/deposit", handlers.Deposit)
	router.POST("/accounts/:id/withdraw", handlers.Withdraw)
	router.POST("/accounts/transfer", handlers.Transfer)

	// Keep original metrics endpoint for compatibility
	router.GET("/metrics", handlers.GetMetrics)
	// Add Prometheus metrics endpoint
	router.GET("/prometheus", handlers.PrometheusMetrics)
	router.GET("/events", handlers.Events)
}
