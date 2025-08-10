package main

import (
	"bank-api/src/diplomat/database"
	"bank-api/src/diplomat/middleware"
	"bank-api/src/diplomat/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	database.Init()
	router := gin.Default()
	router.Use(middleware.CORS())
	routes.RegisterRoutes(router)
	router.Run(":8080")
}
