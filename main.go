package main

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/httpServer"
	"ePrometna_Server/service"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	app.Setup()
	app.Provide(func() *gorm.DB {
		db, err := gorm.Open(postgres.Open(config.AppConfig.DbConnection), &gorm.Config{
			// NOTE: change LogMode if needed when debugging
			Logger: app.NewGormZapLogger().LogMode(logger.Warn),
		})
		if err != nil {
			zap.S().Panicf("failed to provide database dependency, err = %+v", err)
		}
		return db
	})

	app.Provide(service.NewTestService)

	zap.S().Infof("Database: http://localhost:8080")
	zap.S().Infof("swagger: http://localhost:8090/swagger/index.html")

	ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL("http://localhost:8090/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1))

	httpServer.Start()
}
