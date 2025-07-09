package wirebox

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
	"github.com/orzkratos/demokratos/demo1kratos/internal/data"
	"github.com/orzkratos/demokratos/demo1kratos/internal/service"
)

type WireBox struct {
	ConfData       *conf.Data
	DataData       *data.Data
	GreeterUsecase *biz.GreeterUsecase
	GreeterService *service.GreeterService
}

func newWireBox(
	confData *conf.Data,
	dataData *data.Data,
	greeterUsecase *biz.GreeterUsecase,
	greeterService *service.GreeterService,
) *WireBox {
	return &WireBox{
		ConfData:       confData,
		DataData:       dataData,
		GreeterUsecase: greeterUsecase,
		GreeterService: greeterService,
	}
}

func NewWireBox(confData *conf.Data, logger log.Logger) (*WireBox, func(), error) {
	return wireApp(confData, logger)
}
