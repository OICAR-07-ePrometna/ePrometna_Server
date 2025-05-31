package app

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type dbProviderFunc func() *gorm.DB

func newDbConn() *gorm.DB {
	db, err := gorm.Open(postgres.Open(config.AppConfig.DbConnection), &gorm.Config{
		// NOTE: change LogMode if needed when debugging
		Logger: NewGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Errorf("failed to connect database err = %+v", err)
		os.Exit(5)
	}
	return db
}

func testDbConn() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file:db?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		zap.S().Errorf("failed to connect database err = %+v", err)
		os.Exit(5)
	}
	return db
}

func dbSetup() {
	var dbConFunc dbProviderFunc
	if config.AppConfig.Env == config.Test {
		dbConFunc = testDbConn
	} else {
		dbConFunc = newDbConn
	}

	db := dbConFunc()

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

	Provide(dbConFunc)
}
