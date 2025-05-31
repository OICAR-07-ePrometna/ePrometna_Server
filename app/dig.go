package app

import (
	"ePrometna_Server/config"
	"sync"

	"go.uber.org/dig"
	"go.uber.org/zap"
)

var digContainer *dig.Container = nil

var once = sync.Once{}

// Test configures app so that it can be used in unit testing
func Test() {
	digContainer = dig.New()
}

func Setup() {
	once.Do(func() {
		setupLogger()
		digContainer = dig.New()
		dbSetup()
	})
}

func setupLogger() {
	if config.AppConfig.Env == config.Dev || config.AppConfig.Env == config.Test {
		err := devLoggerSetup()
		if err != nil {
			zap.S().Panicf("failed to set up logger, err = %+v", err)
		}
	} else {
		err := prodLoggerSetup()
		if err != nil {
			zap.S().Panicf("failed to set up logger, err = %+v", err)
		}
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
