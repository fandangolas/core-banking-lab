package testenv

import (
	"bank-api/src/db"
	"bank-api/src/routes"
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
		db.Init()
		routes.RegisterRoutes(router)
	})
	return router
}
