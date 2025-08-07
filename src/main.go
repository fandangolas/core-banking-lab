package main

import (
	"bank-api/src/db"
	"bank-api/src/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()
	router := gin.Default()
	routes.RegisterRoutes(router)
	router.Run(":8080")
}
