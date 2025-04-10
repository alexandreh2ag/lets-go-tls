package main

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/cli"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"os"
	"time"
)

func main() {
	expire := time.Date(2025, 04, 25, 13, 00, 00, 00, time.UTC)
	dueDate := time.Now().Add(time.Hour * 24 * 30 * -1)

	fmt.Println("expire", expire)
	fmt.Println("due date", dueDate)
	fmt.Println(expire.After(dueDate))
	os.Exit(0)
	ctx := context.DefaultContext()
	rootCmd := cli.GetRootCmd(ctx)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
