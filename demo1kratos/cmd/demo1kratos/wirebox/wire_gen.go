// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package wirebox

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
	"github.com/orzkratos/demokratos/demo1kratos/internal/data"
	"github.com/orzkratos/demokratos/demo1kratos/internal/service"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confData *conf.Data, logger log.Logger) (*WireBox, func(), error) {
	dataData, cleanup, err := data.NewData(confData, logger)
	if err != nil {
		return nil, nil, err
	}
	greeterRepo := data.NewGreeterRepo(dataData, logger)
	greeterUsecase := biz.NewGreeterUsecase(greeterRepo, logger)
	greeterService := service.NewGreeterService(greeterUsecase)
	wireBox := newWireBox(confData, dataData, greeterRepo, greeterUsecase, greeterService)
	return wireBox, func() {
		cleanup()
	}, nil
}
