package routes

import (
	"bank-api/internal/api/handlers"
	"bank-api/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes with the container dependencies
func RegisterRoutes(router *gin.Engine, container handlers.HandlerDependencies) {
	router.Use(middleware.RequestContextMiddleware()) // Add request-scoped context (first!)
	router.Use(middleware.Metrics())
	router.Use(middleware.PrometheusMiddleware()) // Add Prometheus metrics collection

	// Banking operations - using closure-based handlers with container dependencies
	router.POST("/accounts", handlers.MakeCreateAccountHandler(container))
	router.GET("/accounts/:id/balance", handlers.MakeGetBalanceHandler(container))
	router.POST("/accounts/:id/deposit", handlers.MakeDepositHandler(container))
	router.POST("/accounts/:id/withdraw", handlers.MakeWithdrawHandler(container))
	router.POST("/accounts/transfer", handlers.MakeTransferHandler(container))

	// System endpoints
	router.GET("/metrics", handlers.GetMetrics)
	router.GET("/prometheus", handlers.PrometheusMetrics)
	router.GET("/events", handlers.Events)
}
