package routes

import (
	"bank-api/src/diplomat/middleware"
	"bank-api/src/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	router.Use(middleware.Metrics())
	router.POST("/accounts", handlers.CreateAccount)

	router.GET("/accounts/:id/balance", handlers.GetBalance)
	router.POST("/accounts/:id/deposit", handlers.Deposit)
	router.POST("/accounts/:id/withdraw", handlers.Withdraw)

	router.POST("/accounts/transfer", handlers.Transfer)
	router.GET("/metrics", handlers.GetMetrics)
}
