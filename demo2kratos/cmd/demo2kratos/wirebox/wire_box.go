package wirebox

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
	"github.com/orzkratos/demokratos/demo2kratos/internal/service"
)

// WireBox 简化版依赖注入容器
type WireBox struct {
	// 业务层
	GreeterUsecase *biz.GreeterUsecase

	// 服务层
	GreeterService *service.GreeterService
}

// newWireBox 创建 WireBox 实例
func newWireBox(
	greeterUsecase *biz.GreeterUsecase,
	greeterService *service.GreeterService,
) *WireBox {
	return &WireBox{
		GreeterUsecase: greeterUsecase,
		GreeterService: greeterService,
	}
}

// NewWireBox 创建完整的依赖注入容器
// 用于命令行工具访问所有组件
func NewWireBox(cfg *conf.Bootstrap, logger log.Logger) (*WireBox, func(), error) {
	return wireBox(cfg.Server, cfg.Data, logger)
}

