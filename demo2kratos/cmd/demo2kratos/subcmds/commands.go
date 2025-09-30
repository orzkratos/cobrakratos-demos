package subcmds

import (
	"context"
	"runtime"

	"github.com/go-kratos/kratos/v2/log"
	v1 "github.com/orzkratos/demokratos/demo2kratos/api/helloworld/v1"
	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/cfgpath"
	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/wirebox"
	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/cfgdata"
	"github.com/spf13/cobra"
	"github.com/yyle88/eroticgo"
	"github.com/yyle88/must"
	"github.com/yyle88/neatjson/neatjsons"
)

// NewVersionCmd 创建版本命令
func NewVersionCmd(serviceName, version string, logger log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			slog := log.NewHelper(logger)

			info := map[string]string{
				"service": serviceName,
				"version": version,
				"config":  cfgpath.ConfigPath,
				"go":      runtime.Version(),
			}

			slog.Infof("Version Info: %s", neatjsons.S(info))
		},
	}
}

// NewConfigCmd 创建配置命令
func NewConfigCmd(logger log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
			slog := log.NewHelper(logger)

			slog.Infof("Config: %s", neatjsons.S(cfg))
		},
	}

	return cmd
}

// NewBizCmd 创建业务层命令组
func NewBizCmd(logger log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "biz",
		Short: "Business operations",
	}

	cmd.AddCommand(NewBizCreateGreeterCmd(logger))
	return cmd
}

// NewBizCreateGreeterCmd 创建greeter的业务命令
func NewBizCreateGreeterCmd(logger log.Logger) *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "create-greeter",
		Short: "Create greeter message",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
			wireBox, cleanup, err := wirebox.NewWireBox(cfg, logger)
			must.Done(err)
			defer cleanup()

			slog := log.NewHelper(logger)

			res, err := wireBox.GreeterUsecase.CreateGreeter(context.Background(), &biz.Greeter{
				Hello: message,
			})
			if err != nil {
				slog.Errorf("Failed to create greeter: %v", err)
				return
			}

			slog.Infof("Created: %s", eroticgo.GREEN.Sprint(res.Hello))
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "Hello from CLI", "Greeter message")
	return cmd
}

// NewServiceCmd 创建服务层命令组
func NewServiceCmd(logger log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Service operations",
	}

	cmd.AddCommand(NewServiceSayHelloCmd(logger))
	return cmd
}

// NewServiceSayHelloCmd 服务层SayHello测试
func NewServiceSayHelloCmd(logger log.Logger) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "say-hello",
		Short: "Test SayHello method",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
			wireBox, cleanup, err := wirebox.NewWireBox(cfg, logger)
			must.Done(err)
			defer cleanup()

			slog := log.NewHelper(logger)

			res, err := wireBox.GreeterService.SayHello(context.Background(), &v1.HelloRequest{
				Name: name,
			})
			if err != nil {
				slog.Errorf("SayHello failed: %v", err)
				return
			}

			slog.Infof("Response: %s", eroticgo.CYAN.Sprint(res.Message))
		},
	}

	cmd.Flags().StringVar(&name, "name", "demo", "Name to greet")
	return cmd
}
