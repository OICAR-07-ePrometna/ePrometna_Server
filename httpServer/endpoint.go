package httpServer

import (
	"ePrometna_Server/controller"
	"ePrometna_Server/docs"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// TODO: this will not be used like this only temporary
func setupHandlers(router *gin.Engine) {
	// TODO: Replace gin default logger with zap
	// router.Use(gin.Recovery())
	api := router.Group("/api")
	api.Use(corsHeader())

	// Basic ping
	helloFunc := gin.HandlerFunc(func(c *gin.Context) {
		c.JSON(http.StatusOK, "Hello")
	})

	api.GET("/", helloFunc)
	docs.SwaggerInfo.BasePath = "/api"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	tc := controller.NewTestController()
	tc.RegisterEndpoints(api)
}
