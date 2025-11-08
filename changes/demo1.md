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

## cmd/demo1kratos/cfgpath/cfg_path.go (+4 -0)

```diff
@@ -0,0 +1,4 @@
+package cfgpath
+
+// ConfigPath is the config path.
+var ConfigPath string
```

## cmd/demo1kratos/main.go (+53 -16)

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
+	"github.com/orzkratos/demokratos/demo1kratos/cmd/demo1kratos/cfgpath"
+	"github.com/orzkratos/demokratos/demo1kratos/cmd/demo1kratos/subcmds"
 	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/pkg/cfgdata"
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
+	fmt.Println("service-name:", Name)
 }
 
 func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
@@ -46,7 +47,6 @@
 }
 
 func main() {
-	flag.Parse()
 	logger := log.With(log.NewStdLogger(os.Stdout),
 		"ts", log.DefaultTimestamp,
 		"caller", log.DefaultCaller,
@@ -56,18 +56,55 @@
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
+	var rootCmd = &cobra.Command{
+		Use:   "demo1kratos",
+		Short: "A Kratos microservice",
+		Run: func(cmd *cobra.Command, args []string) {
+			//避免还有其它命令以避免是拼写错误
+			mustslice.None(args)
+			//控制是否在根命令下运行逻辑，当然，直接删掉逻辑也能控制不执行，这里只是给个样例
+			if cfg := cfgdata.ParseConfig(cfgpath.ConfigPath); cfg.Server.AutoRun {
+				runApp(cfg, logger)
+			}
+		},
+	}
+	// 设置全局参数 所有子命令都可以用
+	rootCmd.PersistentFlags().StringVarP(&cfgpath.ConfigPath, "conf", "c", "./configs", "config path, eg: --conf=config.yaml")
 
-	var cfg conf.Bootstrap
-	must.Done(c.Scan(&cfg))
+	// 默认命令（运行服务）
+	rootCmd.AddCommand(&cobra.Command{
+		Use:   "run",
+		Short: "Start the application",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
+			runApp(cfg, logger)
+		},
+	})
 
+	// 在该文件里，添加 version 子命令，读取版本信息
+	// ./bin/demo1kratos version
+	rootCmd.AddCommand(&cobra.Command{
+		Use:   "version",
+		Short: "Print the version info",
+		Run: func(cmd *cobra.Command, args []string) {
+			LOG := log.NewHelper(logger)
+			LOG.Infof("service-name: %s version: %s", Name, Version)
+		},
+	})
+
+	// 在其它文件里添加子命令，这些只是样例，因此随便写写，实际上需要的可能是 migrate 和其它操作，在其它样例里写
+	// ./bin/demo1kratos config
+	rootCmd.AddCommand(subcmds.NewShowConfigCmd(logger))
+	// ./bin/demo1kratos biz create-greeter --message=abc
+	rootCmd.AddCommand(subcmds.NewBizCmd(logger))
+	// ./bin/demo1kratos svc say-hello --name=xyz
+	rootCmd.AddCommand(subcmds.NewSvcCmd(logger))
+	// 执行命令行
+	must.Done(rootCmd.Execute())
+}
+
+func runApp(cfg *conf.Bootstrap, logger log.Logger) {
 	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
 	defer cleanup()
 
```

## cmd/demo1kratos/subcmds/sub_cmds.go (+94 -0)

```diff
@@ -0,0 +1,94 @@
+package subcmds
+
+import (
+	"context"
+
+	"github.com/go-kratos/kratos/v2/log"
+	v1 "github.com/orzkratos/demokratos/demo1kratos/api/helloworld/v1"
+	"github.com/orzkratos/demokratos/demo1kratos/cmd/demo1kratos/cfgpath"
+	"github.com/orzkratos/demokratos/demo1kratos/cmd/demo1kratos/wirebox"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/pkg/cfgdata"
+	"github.com/spf13/cobra"
+	"github.com/yyle88/eroticgo"
+	"github.com/yyle88/must"
+	"github.com/yyle88/neatjson/neatjsons"
+)
+
+func NewShowConfigCmd(logger log.Logger) *cobra.Command {
+	return &cobra.Command{
+		Use:   "config",
+		Short: "Print the config data",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
+			LOG := log.NewHelper(logger)
+			LOG.Infof("Config: %s", neatjsons.S(cfg))
+		},
+	}
+}
+
+func NewBizCmd(logger log.Logger) *cobra.Command {
+	cmd := &cobra.Command{
+		Use:   "biz",
+		Short: "biz",
+	}
+	cmd.AddCommand(NewBizGreaterCmd(logger))
+	return cmd
+}
+
+func NewBizGreaterCmd(logger log.Logger) *cobra.Command {
+	var message string
+	cmd := &cobra.Command{
+		Use:   "create-greeter",
+		Short: "create greeter message",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
+			wireBox, cleanup, err := wirebox.NewWireBox(cfg.Data, logger)
+			must.Done(err)
+			defer cleanup()
+			res, err := wireBox.GreeterUsecase.CreateGreeter(context.Background(), &biz.Greeter{
+				Hello: message,
+			})
+			if err != nil {
+				panic(err)
+			}
+			LOG := log.NewHelper(logger)
+			LOG.Infof("result:%s", eroticgo.GREEN.Sprint(res.Hello))
+		},
+	}
+	cmd.Flags().StringVarP(&message, "message", "", "", "message param")
+	return cmd
+}
+
+func NewSvcCmd(logger log.Logger) *cobra.Command {
+	cmd := &cobra.Command{
+		Use:   "svc",
+		Short: "svc",
+	}
+	cmd.AddCommand(NewSvcGreaterCmd(logger))
+	return cmd
+}
+
+func NewSvcGreaterCmd(logger log.Logger) *cobra.Command {
+	var name string
+	cmd := &cobra.Command{
+		Use:   "say-hello",
+		Short: "say hello to name",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := cfgdata.ParseConfig(cfgpath.ConfigPath)
+			wireBox, cleanup, err := wirebox.NewWireBox(cfg.Data, logger)
+			must.Done(err)
+			defer cleanup()
+			res, err := wireBox.GreeterService.SayHello(context.Background(), &v1.HelloRequest{
+				Name: name,
+			})
+			if err != nil {
+				panic(err)
+			}
+			LOG := log.NewHelper(logger)
+			LOG.Infof("result:%s", eroticgo.AMBER.Sprint(res.Message))
+		},
+	}
+	cmd.Flags().StringVarP(&name, "name", "", "", "name param")
+	return cmd
+}
```

## cmd/demo1kratos/wirebox/wire.go (+25 -0)

```diff
@@ -0,0 +1,25 @@
+//go:build wireinject
+
+// The build tag makes sure the stub is not built in the final build.
+
+package wirebox
+
+import (
+	"github.com/go-kratos/kratos/v2/log"
+	"github.com/google/wire"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/data"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/service"
+)
+
+// wireApp init kratos application.
+func wireApp(*conf.Data, log.Logger) (*WireBox, func(), error) {
+	panic(wire.Build(
+		// server.ProviderSet, // unused provider set "ProviderSet" //需要把没有用到的注释掉
+		data.ProviderSet,
+		biz.ProviderSet,
+		service.ProviderSet,
+		newWireBox,
+	))
+}
```

## cmd/demo1kratos/wirebox/wire_box.go (+37 -0)

```diff
@@ -0,0 +1,37 @@
+package wirebox
+
+import (
+	"github.com/go-kratos/kratos/v2/log"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/data"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/service"
+)
+
+type WireBox struct {
+	ConfData       *conf.Data
+	DataData       *data.Data
+	GreeterRepo    biz.GreeterRepo
+	GreeterUsecase *biz.GreeterUsecase
+	GreeterService *service.GreeterService
+}
+
+func newWireBox(
+	confData *conf.Data,
+	dataData *data.Data,
+	greeterRepo biz.GreeterRepo,
+	greeterUsecase *biz.GreeterUsecase,
+	greeterService *service.GreeterService,
+) *WireBox {
+	return &WireBox{
+		ConfData:       confData,
+		DataData:       dataData,
+		GreeterRepo:    greeterRepo,
+		GreeterUsecase: greeterUsecase,
+		GreeterService: greeterService,
+	}
+}
+
+func NewWireBox(confData *conf.Data, logger log.Logger) (*WireBox, func(), error) {
+	return wireApp(confData, logger)
+}
```

## cmd/demo1kratos/wirebox/wire_gen.go (+31 -0)

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
+	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/data"
+	"github.com/orzkratos/demokratos/demo1kratos/internal/service"
+)
+
+// Injectors from wire.go:
+
+// wireApp init kratos application.
+func wireApp(confData *conf.Data, logger log.Logger) (*WireBox, func(), error) {
+	dataData, cleanup, err := data.NewData(confData, logger)
+	if err != nil {
+		return nil, nil, err
+	}
+	greeterRepo := data.NewGreeterRepo(dataData, logger)
+	greeterUsecase := biz.NewGreeterUsecase(greeterRepo, logger)
+	greeterService := service.NewGreeterService(greeterUsecase)
+	wireBox := newWireBox(confData, dataData, greeterRepo, greeterUsecase, greeterService)
+	return wireBox, func() {
+		cleanup()
+	}, nil
+}
```

## configs/config.yaml (+1 -0)

```diff
@@ -5,6 +5,7 @@
   grpc:
     addr: 0.0.0.0:19000
     timeout: 1s
+  auto_run: true
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
+	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
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

