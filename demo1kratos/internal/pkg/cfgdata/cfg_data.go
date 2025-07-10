package cfgdata

import (
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
	"github.com/yyle88/rese"
)

func ParseConfig(configPath string) *conf.Bootstrap {
	c := config.New(
		config.WithSource(
			file.NewSource(configPath),
		),
	)
	defer rese.F0(c.Close)

	if err := c.Load(); err != nil {
		panic(err)
	}

	var cfg conf.Bootstrap
	if err := c.Scan(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
