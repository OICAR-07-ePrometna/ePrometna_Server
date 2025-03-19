package main

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/httpServer"
	"ePrometna_Server/model"

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

	if config.AppConfig.IsDevelopment {
		devLoggerSetup()
	} else {
		prodLoggerSetup()
	}

	// TODO: register databse as a singleton in app package and lock with mutexes mby rw once
	db, err := gorm.Open(postgres.Open(config.AppConfig.DbConnection), &gorm.Config{
		// NOTE: change LogMode if needed when debugging
		Logger: NewGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().DPanicf("failed to connect database err = %+v", err)
	}

	if err = db.AutoMigrate(model.GetAllModels()...); err != nil {
		zap.S().Panicf("Can't run AutoMigrate err = %+v", err)
	}
	app.Setup()
	// BUG: need to provide constructor function not object
	app.Provide(func() *gorm.DB {
		db, err := gorm.Open(postgres.Open(config.AppConfig.DbConnection), &gorm.Config{
			// NOTE: change LogMode if needed when debugging
			Logger: NewGormZapLogger().LogMode(logger.Warn),
		})
		if err != nil {
			zap.S().Panicf("failed to provide database dependency, err = %+v", err)
		}
		return db
	})

	httpServer.Start()
}
