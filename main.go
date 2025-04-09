package main

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/httpServer"
	"ePrometna_Server/service"

	"go.uber.org/zap"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	app.Setup()

	// Provided logger
	app.Provide(zap.S)

	app.Provide(service.NewTestService)
	app.Provide(service.NewLoginService)
	app.Provide(service.NewUserCrudService)
	app.Provide(service.NewVehicleService)

	zap.S().Infof("Database: http://localhost:8080")
	zap.S().Infof("swagger: http://localhost:8090/swagger/index.html")

	httpServer.Start()
}
