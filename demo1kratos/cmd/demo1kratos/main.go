package main

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/orzkratos/demokratos/demo1kratos/cmd/demo1kratos/cfgpath"
	"github.com/orzkratos/demokratos/demo1kratos/cmd/demo1kratos/subcmds"
	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
	"github.com/orzkratos/demokratos/demo1kratos/internal/pkg/cfgdata"
	"github.com/spf13/cobra"
	"github.com/yyle88/done"
	"github.com/yyle88/must"
	"github.com/yyle88/must/mustslice"
	"github.com/yyle88/rese"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
)

func init() {
	fmt.Println("service-name:", Name)
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
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", kratos.ID(done.VCE(os.Hostname()).Omit()),
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	var rootCmd = &cobra.Command{
		Use:   "demo1kratos",
		Short: "A Kratos microservice",
		Run: func(cmd *cobra.Command, args []string) {
			//避免还有其它命令以避免是拼写错误
			mustslice.None(args)
			//控制是否在根命令下运行逻辑，当然，直接删掉逻辑也能控制不执行，这里只是给个样例
			if cfg := cfgdata.ParseConfig(cfgpath.ConfigPath); cfg.Server.AutoRun {
				runApp(cfg, logger)
			}
		},
	}
	// 设置全局参数 所有子命令都可以用
	rootCmd.PersistentFlags().StringVarP(&cfgpath.ConfigPath, "conf", "c", "./configs", "config path, eg: --conf=config.yaml")

	// 默认命令（运行服务）
	rootCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Start the application",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
			runApp(cfg, logger)
		},
	})

	// 在该文件里，添加 version 子命令，读取版本信息
	// ./bin/demo1kratos version
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version info",
		Run: func(cmd *cobra.Command, args []string) {
			LOG := log.NewHelper(logger)
			LOG.Infof("service-name: %s version: %s", Name, Version)
		},
	})

	// 在其它文件里添加子命令，这些只是样例，因此随便写写，实际上需要的可能是 migrate 和其它操作，在其它样例里写
	// ./bin/demo1kratos config
	rootCmd.AddCommand(subcmds.NewShowConfigCmd(logger))
	// ./bin/demo1kratos biz create-greeter --message=abc
	rootCmd.AddCommand(subcmds.NewBizCmd(logger))
	// ./bin/demo1kratos svc say-hello --name=xyz
	rootCmd.AddCommand(subcmds.NewSvcCmd(logger))
	// 执行命令行
	must.Done(rootCmd.Execute())
}

func runApp(cfg *conf.Bootstrap, logger log.Logger) {
	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
	defer cleanup()

	// start and wait for stop signal
	must.Done(app.Run())
}
