package app

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"

	"go.uber.org/dig"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var digContainer *dig.Container = nil

func Setup() {
	if digContainer == nil {

		if config.AppConfig.IsDevelopment {
			devLoggerSetup()
		} else {
			prodLoggerSetup()
		}

		setupLogger()
		dbSetup()

		digContainer = dig.New()
		return
	}
	zap.S().Warn("app setup is called more than once")
}

func setupLogger() {
}

func dbSetup() {
	db, err := gorm.Open(postgres.Open(config.AppConfig.DbConnection), &gorm.Config{
		// NOTE: change LogMode if needed when debugging
		Logger: NewGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}

	if err = db.AutoMigrate(model.GetAllModels()...); err != nil {
		zap.S().Panicf("Can't run AutoMigrate err = %+v", err)
	}
}

func Provide(service any, opts ...dig.ProvideOption) {
	if err := digContainer.Provide(service, opts...); err != nil {
		zap.S().Panicf("Faild to provide service %T, err = %+v", service, err)
	}
}

func Invoke(service any, opts ...dig.InvokeOption) {
	if err := digContainer.Invoke(service, opts...); err != nil {
		zap.S().Panicf("Faild to provide service %T, err = %+v", service, err)
	}
}
