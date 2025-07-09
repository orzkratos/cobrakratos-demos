//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package wirebox

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
	"github.com/orzkratos/demokratos/demo1kratos/internal/data"
	"github.com/orzkratos/demokratos/demo1kratos/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Data, log.Logger) (*WireBox, func(), error) {
	panic(wire.Build(
		// server.ProviderSet, // unused provider set "ProviderSet" //需要把没有用到的注释掉
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newWireBox,
	))
}
