package httpServer

import (
	"ePrometna_Server/controller"
	"ePrometna_Server/docs"
	"ePrometna_Server/util/middleware"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupHandlers(router *gin.Engine) {
	// TODO: Replace gin default logger with zap
	// router.Use(gin.Recovery())

	router.Use(middleware.CorsHeader())
	api := router.Group("/api")

	// register swagger
	docs.SwaggerInfo.BasePath = "/api"

	swagger := ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL("http://localhost:8090/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(2))
	router.GET("/swagger/*any", swagger)

	controller.NewLoginController().RegisterEndpoints(api)
	controller.NewUserController().RegisterEndpoints(api)
	controller.NewVehicleController().RegisterEndpoints(api)
	controller.NewLicenseController().RegisterEndpoints(api)
}
