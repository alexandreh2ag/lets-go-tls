package cli

import (
	stdContext "context"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	appAgentHttp "github.com/alexandreh2ag/lets-go-tls/apps/agent/http"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/service"
	appHttp "github.com/alexandreh2ag/lets-go-tls/http"
	"github.com/spf13/cobra"
)

func GetStartCmd(ctx *context.AgentContext) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start server",
		RunE:  GetStartRunFn(ctx),
	}
}

func GetStartRunFn(ctx *context.AgentContext) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {

		e := appAgentHttp.CreateServerHTTP(ctx)

		httpConfig := ctx.Config.HTTP
		go appHttp.StartServerHTTP(e, httpConfig.Listen, nil)

		if httpConfig.TLS.Enable {
			tlsConfig := appHttp.CreateTLSConfig(httpConfig.TLS)
			go appHttp.StartServerHTTP(e, httpConfig.TLS.Listen, tlsConfig)
		}

		srv := service.NewService(ctx)
		go func() {
			errStart := srv.Start(ctx)
			if errStart != nil {
				panic(errStart)
			}
		}()

		for {
			select {
			case sig := <-ctx.Signal():
				ctx.Cancel()
				ctx.Logger.Info(fmt.Sprintf("%s signal received, exiting...", sig.String()))
				return e.Shutdown(stdContext.Background())
			}
		}

	}
}
