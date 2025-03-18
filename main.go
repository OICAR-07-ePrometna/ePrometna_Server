package main

import (
	"ePrometna_Server/config"
	"ePrometna_Server/httpServer"
	"ePrometna_Server/model"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	dsn := "host=localhost user=postgres password=postgres dbname=eprometna port=5332 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		zap.S().DPanicf("failed to connect database err = %+v", err)
	}
	db.AutoMigrate(&model.Tmodel{})
	db.Create(&model.Tmodel{Name: "Test insert"})

	httpServer.Start()
}
