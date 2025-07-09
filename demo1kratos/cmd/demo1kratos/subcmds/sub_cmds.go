package subcmds

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "github.com/orzkratos/demokratos/demo1kratos/api/helloworld/v1"
	"github.com/orzkratos/demokratos/demo1kratos/cmd/demo1kratos/cfgdata"
	"github.com/orzkratos/demokratos/demo1kratos/cmd/demo1kratos/wirebox"
	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
	"github.com/spf13/cobra"
	"github.com/yyle88/eroticgo"
	"github.com/yyle88/must"
	"github.com/yyle88/neatjson/neatjsons"
)

func NewShowConfigCmd(logger log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print the config data",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgdata.ParseConfig()
			LOG := log.NewHelper(logger)
			LOG.Infof("Config: %s", neatjsons.S(cfg))
		},
	}
}

func NewBizCmd(logger log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "biz",
		Short: "biz",
	}
	cmd.AddCommand(NewBizGreaterCmd(logger))
	return cmd
}

func NewBizGreaterCmd(logger log.Logger) *cobra.Command {
	var message string
	cmd := &cobra.Command{
		Use:   "create-greeter",
		Short: "create greeter message",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgdata.ParseConfig()
			wireBox, cleanup, err := wirebox.NewWireBox(cfg.Data, logger)
			must.Done(err)
			defer cleanup()
			res, err := wireBox.GreeterUsecase.CreateGreeter(context.Background(), &biz.Greeter{
				Hello: message,
			})
			if err != nil {
				panic(err)
			}
			LOG := log.NewHelper(logger)
			LOG.Infof("result:%s", eroticgo.GREEN.Sprint(res.Hello))
		},
	}
	cmd.Flags().StringVarP(&message, "message", "", "", "message param")
	return cmd
}

func NewSvcCmd(logger log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "svc",
		Short: "svc",
	}
	cmd.AddCommand(NewSvcGreaterCmd(logger))
	return cmd
}

func NewSvcGreaterCmd(logger log.Logger) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "say-hello",
		Short: "say hello to name",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := cfgdata.ParseConfig()
			wireBox, cleanup, err := wirebox.NewWireBox(cfg.Data, logger)
			must.Done(err)
			defer cleanup()
			res, err := wireBox.GreeterService.SayHello(context.Background(), &v1.HelloRequest{
				Name: name,
			})
			if err != nil {
				panic(err)
			}
			LOG := log.NewHelper(logger)
			LOG.Infof("result:%s", eroticgo.AMBER.Sprint(res.Message))
		},
	}
	cmd.Flags().StringVarP(&name, "name", "", "", "name param")
	return cmd
}
