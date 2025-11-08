# Changes

Code differences compared to source project demokratos.

## Makefile (+1 -0)

```diff
@@ -66,6 +66,7 @@
 	go mod tidy
 	@# Remove obsolete +build tags from wire_gen.go
 	@sed -i '' '/^\/\/ +build/d' ./cmd/*/wire_gen.go
+	@sed -i '' '/^\/\/ +build/d' ./cmd/*/wirebox/wire_gen.go
 
 .PHONY: all
 # generate all
```

## cmd/demo2kratos/cfgpath/cfg_path.go (+4 -0)

```diff
@@ -0,0 +1,4 @@
+package cfgpath
+
+// ConfigPath is the config path.
+var ConfigPath string
```

## cmd/demo2kratos/main.go (+53 -16)

```diff
@@ -1,19 +1,22 @@
 package main
 
 import (
-	"flag"
+	"fmt"
 	"os"
 
 	"github.com/go-kratos/kratos/v2"
-	"github.com/go-kratos/kratos/v2/config"
-	"github.com/go-kratos/kratos/v2/config/file"
 	"github.com/go-kratos/kratos/v2/log"
 	"github.com/go-kratos/kratos/v2/middleware/tracing"
 	"github.com/go-kratos/kratos/v2/transport/grpc"
 	"github.com/go-kratos/kratos/v2/transport/http"
+	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/cfgpath"
+	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/subcmds"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/cfgdata"
+	"github.com/spf13/cobra"
 	"github.com/yyle88/done"
 	"github.com/yyle88/must"
+	"github.com/yyle88/must/mustslice"
 	"github.com/yyle88/rese"
 )
 
@@ -23,12 +26,10 @@
 	Name string
 	// Version is the version of the compiled software.
 	Version string
-	// flagconf is the config flag.
-	flagconf string
 )
 
 func init() {
-	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
+	fmt.Println("Initializing", Name, Version)
 }
 
 func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
@@ -46,7 +47,7 @@
 }
 
 func main() {
-	flag.Parse()
+	// 创建日志实例
 	logger := log.With(log.NewStdLogger(os.Stdout),
 		"ts", log.DefaultTimestamp,
 		"caller", log.DefaultCaller,
@@ -56,18 +57,54 @@
 		"trace.id", tracing.TraceID(),
 		"span.id", tracing.SpanID(),
 	)
-	c := config.New(
-		config.WithSource(
-			file.NewSource(flagconf),
-		),
-	)
-	defer rese.F0(c.Close)
 
-	must.Done(c.Load())
 
-	var cfg conf.Bootstrap
-	must.Done(c.Scan(&cfg))
+	// 创建根命令
+	rootCmd := &cobra.Command{
+		Use:   Name,
+		Short: "Kratos microservice with Cobra CLI",
+		Long:  fmt.Sprintf("%s - Kratos microservice demonstration", Name),
+		Run: func(cmd *cobra.Command, args []string) {
+			// 避免参数错误
+			mustslice.None(args)
 
+			// 检查是否应该自动运行服务
+			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
+			if cfg.Server.AutoRun {
+				runApp(cfg, logger)
+			} else {
+				fmt.Printf("Service configured not to auto-run. Use '%s run' to start.\n", Name)
+				fmt.Printf("Use '%s --help' to see available commands.\n", Name)
+			}
+		},
+	}
+
+	// 全局参数
+	rootCmd.PersistentFlags().StringVarP(&cfgpath.ConfigPath, "config", "c", "./configs/config.yaml", "Config file path")
+
+	// 核心命令
+	rootCmd.AddCommand(&cobra.Command{
+		Use:   "run",
+		Short: "Start the service",
+		Long:  "Start the HTTP and gRPC servers",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
+			runApp(cfg, logger)
+		},
+	})
+
+	// 添加子命令
+	rootCmd.AddCommand(subcmds.NewVersionCmd(Name, Version, logger))
+	rootCmd.AddCommand(subcmds.NewConfigCmd(logger))
+	rootCmd.AddCommand(subcmds.NewBizCmd(logger))
+	rootCmd.AddCommand(subcmds.NewServiceCmd(logger))
+
+	// 执行命令
+	must.Done(rootCmd.Execute())
+}
+
+// runApp 启动应用服务
+func runApp(cfg *conf.Bootstrap, logger log.Logger) {
 	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
 	defer cleanup()
 
```

## cmd/demo2kratos/subcmds/commands.go (+137 -0)

```diff
@@ -0,0 +1,137 @@
+package subcmds
+
+import (
+	"context"
+	"runtime"
+
+	"github.com/go-kratos/kratos/v2/log"
+	v1 "github.com/orzkratos/demokratos/demo2kratos/api/helloworld/v1"
+	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/cfgpath"
+	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/wirebox"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/cfgdata"
+	"github.com/spf13/cobra"
+	"github.com/yyle88/eroticgo"
+	"github.com/yyle88/must"
+	"github.com/yyle88/neatjson/neatjsons"
+)
+
+// NewVersionCmd 创建版本命令
+func NewVersionCmd(serviceName, version string, logger log.Logger) *cobra.Command {
+	return &cobra.Command{
+		Use:   "version",
+		Short: "Print version info",
+		Run: func(cmd *cobra.Command, args []string) {
+			slog := log.NewHelper(logger)
+
+			info := map[string]string{
+				"service": serviceName,
+				"version": version,
+				"config":  cfgpath.ConfigPath,
+				"go":      runtime.Version(),
+			}
+
+			slog.Infof("Version Info: %s", neatjsons.S(info))
+		},
+	}
+}
+
+// NewConfigCmd 创建配置命令
+func NewConfigCmd(logger log.Logger) *cobra.Command {
+	cmd := &cobra.Command{
+		Use:   "config",
+		Short: "Show configuration",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
+			slog := log.NewHelper(logger)
+
+			slog.Infof("Config: %s", neatjsons.S(cfg))
+		},
+	}
+
+	return cmd
+}
+
+// NewBizCmd 创建业务层命令组
+func NewBizCmd(logger log.Logger) *cobra.Command {
+	cmd := &cobra.Command{
+		Use:   "biz",
+		Short: "Business operations",
+	}
+
+	cmd.AddCommand(NewBizCreateGreeterCmd(logger))
+	return cmd
+}
+
+// NewBizCreateGreeterCmd 创建greeter的业务命令
+func NewBizCreateGreeterCmd(logger log.Logger) *cobra.Command {
+	var message string
+
+	cmd := &cobra.Command{
+		Use:   "create-greeter",
+		Short: "Create greeter message",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
+			wireBox, cleanup, err := wirebox.NewWireBox(cfg, logger)
+			must.Done(err)
+			defer cleanup()
+
+			slog := log.NewHelper(logger)
+
+			res, err := wireBox.GreeterUsecase.CreateGreeter(context.Background(), &biz.Greeter{
+				Hello: message,
+			})
+			if err != nil {
+				slog.Errorf("Failed to create greeter: %v", err)
+				return
+			}
+
+			slog.Infof("Created: %s", eroticgo.GREEN.Sprint(res.Hello))
+		},
+	}
+
+	cmd.Flags().StringVarP(&message, "message", "m", "Hello from CLI", "Greeter message")
+	return cmd
+}
+
+// NewServiceCmd 创建服务层命令组
+func NewServiceCmd(logger log.Logger) *cobra.Command {
+	cmd := &cobra.Command{
+		Use:   "service",
+		Short: "Service operations",
+	}
+
+	cmd.AddCommand(NewServiceSayHelloCmd(logger))
+	return cmd
+}
+
+// NewServiceSayHelloCmd 服务层SayHello测试
+func NewServiceSayHelloCmd(logger log.Logger) *cobra.Command {
+	var name string
+
+	cmd := &cobra.Command{
+		Use:   "say-hello",
+		Short: "Test SayHello method",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
+			wireBox, cleanup, err := wirebox.NewWireBox(cfg, logger)
+			must.Done(err)
+			defer cleanup()
+
+			slog := log.NewHelper(logger)
+
+			res, err := wireBox.GreeterService.SayHello(context.Background(), &v1.HelloRequest{
+				Name: name,
+			})
+			if err != nil {
+				slog.Errorf("SayHello failed: %v", err)
+				return
+			}
+
+			slog.Infof("Response: %s", eroticgo.CYAN.Sprint(res.Message))
+		},
+	}
+
+	cmd.Flags().StringVar(&name, "name", "demo", "Name to greet")
+	return cmd
+}
```

## cmd/demo2kratos/wirebox/wire.go (+22 -0)

```diff
@@ -0,0 +1,22 @@
+//go:build wireinject
+
+package wirebox
+
+import (
+	"github.com/go-kratos/kratos/v2/log"
+	"github.com/google/wire"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/data"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/service"
+)
+
+// wireBox 简化的依赖注入容器
+func wireBox(*conf.Server, *conf.Data, log.Logger) (*WireBox, func(), error) {
+	panic(wire.Build(
+		data.ProviderSet,
+		biz.ProviderSet,
+		service.ProviderSet,
+		newWireBox,
+	))
+}
```

## cmd/demo2kratos/wirebox/wire_box.go (+35 -0)

```diff
@@ -0,0 +1,35 @@
+package wirebox
+
+import (
+	"github.com/go-kratos/kratos/v2/log"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/service"
+)
+
+// WireBox 简化版依赖注入容器
+type WireBox struct {
+	// 业务层
+	GreeterUsecase *biz.GreeterUsecase
+
+	// 服务层
+	GreeterService *service.GreeterService
+}
+
+// newWireBox 创建 WireBox 实例
+func newWireBox(
+	greeterUsecase *biz.GreeterUsecase,
+	greeterService *service.GreeterService,
+) *WireBox {
+	return &WireBox{
+		GreeterUsecase: greeterUsecase,
+		GreeterService: greeterService,
+	}
+}
+
+// NewWireBox 创建完整的依赖注入容器
+// 用于命令行工具访问所有组件
+func NewWireBox(cfg *conf.Bootstrap, logger log.Logger) (*WireBox, func(), error) {
+	return wireBox(cfg.Server, cfg.Data, logger)
+}
+
```

## cmd/demo2kratos/wirebox/wire_gen.go (+31 -0)

```diff
@@ -0,0 +1,31 @@
+// Code generated by Wire. DO NOT EDIT.
+
+//go:generate go run -mod=mod github.com/google/wire/cmd/wire
+//go:build !wireinject
+
+package wirebox
+
+import (
+	"github.com/go-kratos/kratos/v2/log"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/data"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/service"
+)
+
+// Injectors from wire.go:
+
+// wireBox 简化的依赖注入容器
+func wireBox(server *conf.Server, confData *conf.Data, logger log.Logger) (*WireBox, func(), error) {
+	dataData, cleanup, err := data.NewData(confData, logger)
+	if err != nil {
+		return nil, nil, err
+	}
+	greeterRepo := data.NewGreeterRepo(dataData, logger)
+	greeterUsecase := biz.NewGreeterUsecase(greeterRepo, logger)
+	greeterService := service.NewGreeterService(greeterUsecase)
+	wireboxWireBox := newWireBox(greeterUsecase, greeterService)
+	return wireboxWireBox, func() {
+		cleanup()
+	}, nil
+}
```

## configs/config.yaml (+1 -0)

```diff
@@ -5,6 +5,7 @@
   grpc:
     addr: 0.0.0.0:29000
     timeout: 1s
+  auto_run: true  # 是否在根命令下自动运行服务
 data:
   database:
     driver: mysql
```

## internal/conf/conf.pb.go (+11 -2)

```diff
@@ -78,6 +78,7 @@
 	state         protoimpl.MessageState `protogen:"open.v1"`
 	Http          *Server_HTTP           `protobuf:"bytes,1,opt,name=http,proto3" json:"http,omitempty"`
 	Grpc          *Server_GRPC           `protobuf:"bytes,2,opt,name=grpc,proto3" json:"grpc,omitempty"`
+	AutoRun       bool                   `protobuf:"varint,3,opt,name=auto_run,json=autoRun,proto3" json:"auto_run,omitempty"`
 	unknownFields protoimpl.UnknownFields
 	sizeCache     protoimpl.SizeCache
 }
@@ -126,6 +127,13 @@
 	return nil
 }
 
+func (x *Server) GetAutoRun() bool {
+	if x != nil {
+		return x.AutoRun
+	}
+	return false
+}
+
 type Data struct {
 	state         protoimpl.MessageState `protogen:"open.v1"`
 	Database      *Data_Database         `protobuf:"bytes,1,opt,name=database,proto3" json:"database,omitempty"`
@@ -426,10 +434,11 @@
 	"kratos.api\x1a\x1egoogle/protobuf/duration.proto\"]\n" +
 	"\tBootstrap\x12*\n" +
 	"\x06server\x18\x01 \x01(\v2\x12.kratos.api.ServerR\x06server\x12$\n" +
-	"\x04data\x18\x02 \x01(\v2\x10.kratos.api.DataR\x04data\"\xb8\x02\n" +
+	"\x04data\x18\x02 \x01(\v2\x10.kratos.api.DataR\x04data\"\xd3\x02\n" +
 	"\x06Server\x12+\n" +
 	"\x04http\x18\x01 \x01(\v2\x17.kratos.api.Server.HTTPR\x04http\x12+\n" +
-	"\x04grpc\x18\x02 \x01(\v2\x17.kratos.api.Server.GRPCR\x04grpc\x1ai\n" +
+	"\x04grpc\x18\x02 \x01(\v2\x17.kratos.api.Server.GRPCR\x04grpc\x12\x19\n" +
+	"\bauto_run\x18\x03 \x01(\bR\aautoRun\x1ai\n" +
 	"\x04HTTP\x12\x18\n" +
 	"\anetwork\x18\x01 \x01(\tR\anetwork\x12\x12\n" +
 	"\x04addr\x18\x02 \x01(\tR\x04addr\x123\n" +
```

## internal/conf/conf.proto (+1 -0)

```diff
@@ -23,6 +23,7 @@
   }
   HTTP http = 1;
   GRPC grpc = 2;
+  bool auto_run = 3;
 }
 
 message Data {
```

## internal/pkg/cfgdata/cfg_data.go (+27 -0)

```diff
@@ -0,0 +1,27 @@
+package cfgdata
+
+import (
+	"github.com/go-kratos/kratos/v2/config"
+	"github.com/go-kratos/kratos/v2/config/file"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/yyle88/rese"
+)
+
+func ParseConfig(configPath string) *conf.Bootstrap {
+	c := config.New(
+		config.WithSource(
+			file.NewSource(configPath),
+		),
+	)
+	defer rese.F0(c.Close)
+
+	if err := c.Load(); err != nil {
+		panic(err)
+	}
+
+	var cfg conf.Bootstrap
+	if err := c.Scan(&cfg); err != nil {
+		panic(err)
+	}
+	return &cfg
+}
```

