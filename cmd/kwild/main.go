package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kwilteam/kwil-db/app"
	"github.com/kwilteam/kwil-db/app/shared"
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

	if err := rootCmd.ExecuteContext(ctx); err != nil { // command syntax error
		os.Exit(-1)
	}

	// For a command / application error, which handle the output themselves, we
	// detect those case where display.PrintErr() is called so that we can
	// return a non-zero exit code, which is important for scripting etc.
	if err := shared.CmdCtxErr(rootCmd); err != nil {
		os.Exit(-1)
	}

	os.Exit(0)
}
