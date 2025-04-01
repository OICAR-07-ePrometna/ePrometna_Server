package httpServer

import (
	"ePrometna_Server/controller"
	"ePrometna_Server/docs"

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

	// register swagger
	docs.SwaggerInfo.BasePath = "/api"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//api.Use(AllowAccess(model.RoleFirma, model.RoleAdmin))

	// testController
	tc := controller.NewTestController()
	lc := controller.NewLoginController()

	tc.RegisterEndpoints(api)
	lc.RegisterEndpoints(api)

	/*
		tp := controller.NewTestController()
		protected := api.Group("/protected")
		protected.Use(protect())
		tp.RegisterEndpoints(protected)
	*/
}
