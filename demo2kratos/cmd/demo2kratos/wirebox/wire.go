//go:build wireinject

package wirebox

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
	"github.com/orzkratos/demokratos/demo2kratos/internal/data"
	"github.com/orzkratos/demokratos/demo2kratos/internal/service"
)

// wireBox 简化的依赖注入容器
func wireBox(*conf.Server, *conf.Data, log.Logger) (*WireBox, func(), error) {
	panic(wire.Build(
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newWireBox,
	))
}
