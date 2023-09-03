package main

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/cli"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
)

func main() {
	ctx := context.DefaultContext()
	rootCmd := cli.GetRootCmd(ctx)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
