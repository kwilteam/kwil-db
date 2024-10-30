package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"kwil/app"

	"github.com/spf13/pflag"
	// TODO: isolate to config package not main
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("shutdown signal received")
		cancel()
	}()

	rootCmd := app.RootCmd()

	// Run "start" as the default command if none is given.
	cmd, _, err := rootCmd.Find(os.Args[1:])
	if err == nil && cmd.Use == rootCmd.Use && cmd.Flags().Parse(os.Args[1:]) != pflag.ErrHelp {
		// rewrite from "kwild <whatever...>" to "kwild start <whatever...>"
		args := append([]string{"start"}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(-1)
	}
}
