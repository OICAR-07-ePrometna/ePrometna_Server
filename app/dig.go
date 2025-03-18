package app

import (
	"go.uber.org/dig"
	"go.uber.org/zap"
)

var digContainer *dig.Container = nil

func Setup() {
	if digContainer == nil {
		digContainer = dig.New()
		return
	}
	zap.S().Warn("app setup is called more than once")
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
