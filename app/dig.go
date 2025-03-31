package app

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"sync"
	"time"

	"go.uber.org/dig"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var digContainer *dig.Container = nil

var once = sync.Once{}

func Setup() {
	once.Do(func() {
		setupLogger()
		digContainer = dig.New()
		dbSetup()
	})
}

func setupLogger() {
	if config.AppConfig.IsDevelopment {
		err := devLoggerSetup()
		if err != nil {
			zap.S().Panicf("failed to connect database err = %+v", err)
		}
	} else {
		err := prodLoggerSetup()
		if err != nil {
			zap.S().Panicf("failed to connect database err = %+v", err)
		}
	}
}

func newDbConn() *gorm.DB {
	db, err := gorm.Open(postgres.Open(config.AppConfig.DbConnection), &gorm.Config{
		// NOTE: change LogMode if needed when debugging
		Logger: NewGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}
	return db
}

func dbSetup() {
	db := newDbConn()
	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Panicf("failed to get database connection: %+v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err = db.AutoMigrate(model.GetAllModels()...); err != nil {
		zap.S().Panicf("Can't run AutoMigrate err = %+v", err)
	}

	Provide(newDbConn)
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
