package cfgdata

import (
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
	"github.com/yyle88/rese"
)

// ConfigPath is the config path.
var ConfigPath string

func ParseConfig() *conf.Bootstrap {
	c := config.New(
		config.WithSource(
			file.NewSource(ConfigPath),
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
