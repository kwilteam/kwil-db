package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/cmd/kwild/server"
	"github.com/kwilteam/kwil-db/internal/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/kwilteam/kwil-db/extensions/auth"
)

var (
	kwildCfg = config.DefaultConfig()
)

func main() {
	rootCmd.Version = version.KwilVersion

	flagSet := rootCmd.Flags()
	flagSet.SortFlags = false
	addKwildFlags(flagSet, kwildCfg)
	viper.BindPFlags(flagSet)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

var rootCmd = &cobra.Command{
	Use:               "kwild",
	Short:             "kwild node and rpc server",
	Long:              "kwild: the Kwil blockchain node and RPC server",
	DisableAutoGenTag: true,
	Args:              cobra.NoArgs, // just flags
	RunE: func(cmd *cobra.Command, args []string) error {
		// command line flags are now (finally) parsed by cobra. Bind env vars,
		// load the config file, and unmarshal into kwildCfg.
		if err := kwildCfg.LoadKwildConfig(); err != nil {
			return fmt.Errorf("failed to load kwild config: %w", err)
		}

		nodeKey, genesisConfig, err := kwildCfg.InitPrivateKeyAndGenesis()
		if err != nil {
			return fmt.Errorf("failed to initialize private key and genesis: %w", err)
		}

		stopProfiler, err := startProfilers(kwildCfg)
		if err != nil {
			return err
		}
		defer stopProfiler()

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(cmd.Context())

		go func() {
			<-signalChan
			cancel()
		}()

		svr, err := server.New(ctx, kwildCfg, genesisConfig, nodeKey)
		if err != nil {
			return err
		}

		return svr.Start(ctx)
	},
}

func startProfilers(cfg *config.KwildConfig) (func(), error) {
	mode := cfg.AppCfg.ProfileMode
	pprofFile := kwildCfg.AppCfg.ProfileFile
	if pprofFile == "" {
		pprofFile = fmt.Sprintf("kwild-%s.pprof", mode)
	}

	switch cfg.AppCfg.ProfileMode {
	case "http":
		// http pprof uses http.DefaultServeMux, so we register a redirect
		// handler with the root path on the default mux.
		http.Handle("/", http.RedirectHandler("/debug/pprof/", http.StatusSeeOther))
		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				fmt.Printf("http.ListenAndServe: %v\n", err)
			}
		}()
		return func() {}, nil
	case "cpu":
		f, err := os.Create(pprofFile)
		if err != nil {
			return nil, err
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return nil, fmt.Errorf("error starting CPU profiler: %w", err)
		}
		return pprof.StopCPUProfile, nil
	case "mem":
		f, err := os.Create(pprofFile)
		if err != nil {
			return nil, err
		}
		timer := time.NewTimer(time.Second * 15)
		go func() {
			<-timer.C
			if err = pprof.WriteHeapProfile(f); err != nil {
				fmt.Printf("WriteHeapProfile: %v\n", err)
			}
			f.Close()
		}()
		return func() { timer.Reset(0) }, nil
	case "block":
		f, err := os.Create(pprofFile)
		if err != nil {
			return nil, fmt.Errorf("could not create block profile file %q: %v", pprofFile, err)
		}
		runtime.SetBlockProfileRate(1)
		return func() {
			pprof.Lookup("block").WriteTo(f, 0)
			f.Close()
			runtime.SetBlockProfileRate(0)
		}, nil
	case "mutex":
		f, err := os.Create(pprofFile)
		if err != nil {
			return nil, fmt.Errorf("could not create mutex profile file %q: %v", pprofFile, err)
		}
		runtime.SetMutexProfileFraction(1)
		return func() {
			if mp := pprof.Lookup("mutex"); mp != nil {
				mp.WriteTo(f, 0)
			}
			f.Close()
			runtime.SetMutexProfileFraction(0)
		}, nil
	case "": // disabled
		return func() {}, nil
	default:
		return nil, fmt.Errorf("unknown profile mode %s", cfg.AppCfg.ProfileMode)
	}
}
