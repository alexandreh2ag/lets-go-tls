package cli

import (
	"github.com/alexandreh2ag/lets-go-tls/version"
	"github.com/spf13/cobra"
)

func GetVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run:   GetVersionRunFn(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func GetVersionRunFn() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		cmd.Println(version.GetFormattedVersion())
	}
}
