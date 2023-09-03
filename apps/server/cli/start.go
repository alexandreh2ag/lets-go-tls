package cli

import (
	stdContext "context"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/http"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/manager"
	"github.com/spf13/cobra"
)

func GetStartCmd(ctx *context.ServerContext) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start server",
		RunE:  GetStartRunFn(ctx),
	}
}

func GetStartRunFn(ctx *context.ServerContext) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {

		e, err := http.CreateServerHTTP(ctx, acme.GetHTTPProvider(ctx))
		if err != nil {
			return err
		}

		mgr, _ := manager.CreateManager(ctx)

		go func() {
			err = mgr.Start(ctx)
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
