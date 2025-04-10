package cli

import (
	"encoding/json"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/migrate"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func GetMigrateCmd(ctx *context.ServerContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate certificates and account from other tools",
		RunE:  GetMigrateRunFn(ctx),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.Flags().StringP("type", "t", "", "Define type migration")
	cmd.Flags().StringP("path", "p", "", "Define source path")
	cmd.Flags().StringP("output", "o", "./state.json", "Define output path for state")
	return cmd
}

func GetMigrateRunFn(ctx *context.ServerContext) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		var state *types.State
		typeMigrate, _ := cmd.Flags().GetString("type")
		sourcePath, _ := cmd.Flags().GetString("path")
		outputPath, _ := cmd.Flags().GetString("output")

		if typeMigrate == "" {
			return fmt.Errorf("type migration is required")
		}

		if sourcePath == "" {
			return fmt.Errorf("source path is required")
		}

		switch typeMigrate {

		case "traefik":
			state, err = migrate.MigrateTraefik(ctx, sourcePath)
		case "certbot":
			state, err = migrate.MigrateCertbot(ctx, sourcePath)
		}

		if err != nil {
			return err
		}
		content, _ := json.Marshal(state)
		return afero.WriteFile(ctx.Fs, outputPath, content, 0644)
	}
}
