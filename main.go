package main

import (
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

	db, err := gorm.Open(postgres.Open(config.AppConfig.DbConnection), &gorm.Config{
		// NOTE: change LogMode if needed when debugging
		Logger: NewGormZapLogger().LogMode(logger.Silent),
	})
	if err != nil {
		zap.S().DPanicf("failed to connect database err = %+v", err)
	}

	if err = db.AutoMigrate(model.GetAllModels()...); err != nil {
		zap.S().Panicf("Can't run AutoMigrate err = %+v", err)
	}

	// TODO: this is test insert remove later
	db.Create(&model.Tmodel{Name: "Test insert"})

	httpServer.Start()
}
