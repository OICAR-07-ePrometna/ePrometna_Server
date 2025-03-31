package main

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/httpServer"
	"ePrometna_Server/service"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	app.Setup()
	app.Provide(service.NewTestService)

	zap.S().Infof("Database: http://localhost:8080")
	zap.S().Infof("swagger: http://localhost:8090/swagger/index.html")

	ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL("http://localhost:8090/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1))

	httpServer.Start()
}
