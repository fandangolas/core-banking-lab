package testenv

import (
	"bank-api/src/diplomat/database"
	"bank-api/src/diplomat/routes"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	router    *gin.Engine
	setupOnce sync.Once
)

func SetupRouter() *gin.Engine {
	setupOnce.Do(func() {
		gin.SetMode(gin.TestMode)
		router = gin.Default()
		database.Init()
		routes.RegisterRoutes(router)
	})
	return router
}
