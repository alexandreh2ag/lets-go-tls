package cli

import (
	stdContext "context"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/http"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/service"
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

		e, err := http.CreateServerHTTP(ctx)
		if err != nil {
			return err
		}

		srv := service.NewService(ctx)
		go func() {
			err = srv.Start(ctx)
			if err != nil {
				panic(err)
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
