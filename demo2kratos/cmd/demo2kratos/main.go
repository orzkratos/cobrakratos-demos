package main

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/cfgpath"
	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/subcmds"
	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/cfgdata"
	"github.com/spf13/cobra"
	"github.com/yyle88/done"
	"github.com/yyle88/must"
	"github.com/yyle88/must/mustslice"
	"github.com/yyle88/rese"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
)

func init() {
	fmt.Println("Initializing", Name, Version)
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(done.VCE(os.Hostname()).Omit()),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	// 创建日志实例
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", kratos.ID(done.VCE(os.Hostname()).Omit()),
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)


	// 创建根命令
	rootCmd := &cobra.Command{
		Use:   Name,
		Short: "Kratos microservice with Cobra CLI",
		Long:  fmt.Sprintf("%s - Kratos microservice demonstration", Name),
		Run: func(cmd *cobra.Command, args []string) {
			// 避免参数错误
			mustslice.None(args)

			// 检查是否应该自动运行服务
			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
			if cfg.Server.AutoRun {
				runApp(cfg, logger)
			} else {
				fmt.Printf("Service configured not to auto-run. Use '%s run' to start.\n", Name)
				fmt.Printf("Use '%s --help' to see available commands.\n", Name)
			}
		},
	}

	// 全局参数
	rootCmd.PersistentFlags().StringVarP(&cfgpath.ConfigPath, "config", "c", "./configs/config.yaml", "Config file path")

	// 核心命令
	rootCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Start the service",
		Long:  "Start the HTTP and gRPC servers",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
			runApp(cfg, logger)
		},
	})

	// 添加子命令
	rootCmd.AddCommand(subcmds.NewVersionCmd(Name, Version, logger))
	rootCmd.AddCommand(subcmds.NewConfigCmd(logger))
	rootCmd.AddCommand(subcmds.NewBizCmd(logger))
	rootCmd.AddCommand(subcmds.NewServiceCmd(logger))

	// 执行命令
	must.Done(rootCmd.Execute())
}

// runApp 启动应用服务
func runApp(cfg *conf.Bootstrap, logger log.Logger) {
	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
	defer cleanup()

	// start and wait for stop signal
	must.Done(app.Run())
}
