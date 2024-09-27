package root

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/kwilteam/kwil-db/cmd"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	kwildcfg "github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/cmd/kwild/server"
	"github.com/kwilteam/kwil-db/core/log"

	// kwild's "root" command package assumes the responsibility of initializing
	// certain packages, including the extensions and the chain config package.
	// After loading the genesis.json file, the global chain.Forks instance is
	// initialized with the hardfork activations defined in the file.
	"github.com/kwilteam/kwil-db/common/chain"
	config "github.com/kwilteam/kwil-db/common/config"
	_ "github.com/kwilteam/kwil-db/extensions" // a base location where all extensions can be registered
	_ "github.com/kwilteam/kwil-db/extensions/auth"
	"github.com/kwilteam/kwil-db/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var long = "`" + `%s` + "`" + ` runs the %s blockchain node and RPC server.

` + "`" + `%s` + "`" + ` is a full-node implementation of the Kwil protocol. It provides the
ability to fully participate in a Kwil network, and can run as either a
validator or a non-validating node.

Extensions can be configured by passing flags delimited by a double dash, e.g.
"%s --autogen -- --extension.extension1.flag1 value1 --extension.extension2.flag2 value2"
This follows the POSIX guidelines for additional operands.
`

var example = `# Start %s and auto-generate a private key, genesis file, and config file
%s --autogen

# Start %s from a root directory with a config file
%s --root-dir /path/to/root

# Start %s with extensions
%s -- --extension.extension1.flag1=value1 --extension.extension2.flag2=value2`

// RootCmd creates a new `kwild` node root command
func RootCmd() *cobra.Command {
	// we use an empty config because this config gets merged later, and should only contain flag values
	flagCfg := kwildcfg.EmptyConfig()
	var autoGen bool

	cmd1 := &cobra.Command{
		Use:               cmd.BinaryConfig.NodeCmd,
		Short:             cmd.BinaryConfig.ProjectName + " node and rpc server",
		Long:              fmt.Sprintf(long, cmd.BinaryConfig.NodeUsage(), cmd.BinaryConfig.ProjectName, cmd.BinaryConfig.NodeUsage(), cmd.BinaryConfig.NodeUsage()),
		Example:           fmt.Sprintf(example, cmd.BinaryConfig.NodeCmd, cmd.BinaryConfig.NodeUsage(), cmd.BinaryConfig.NodeCmd, cmd.BinaryConfig.NodeUsage(), cmd.BinaryConfig.NodeCmd, cmd.BinaryConfig.NodeUsage()),
		DisableAutoGenTag: true,
		Version:           version.KwilVersion,
		SilenceUsage:      true, // not all errors imply cli misuse
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("kwild version %v (Go version %s)\n", version.KwilVersion, runtime.Version())

			// args can be passed to kwild in the form of extension flags.
			// These are delimited using a double dash, e.g. `kwild -- --extension.extension1.flag1=value1 --extension.extension2.flag2=value2`
			// This is in-line with guideline 10 of the POSIX guidelines: https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap12.html#tag_12_02
			extensionConfig, err := parseExtensionFlags(args)
			if err != nil {
				return fmt.Errorf("failed to parse extension flags: %w", err)
			}
			flagCfg.AppConfig.Extensions = extensionConfig

			kwildCfg, configFileExists, err := kwildcfg.GetCfg(flagCfg)
			if err != nil {
				cmd.Usage()
				return err
			}

			nodeKey, genesisConfig, err := kwildcfg.InitPrivateKeyAndGenesis(kwildCfg, autoGen)
			if err != nil {
				return fmt.Errorf("failed to initialize private key and genesis: %w", err)
			}

			logCfg, err := kwildCfg.LogConfig()
			if err != nil {
				return err
			}

			logger, err := log.NewChecked(*logCfg)
			if err != nil {
				return fmt.Errorf("invalid logger config: %w", err)
			}
			logger = *logger.Named("kwild")

			logger.Infof("version %s commit %s", version.KwilVersion, version.Build.RevisionShort)

			if kwildCfg.MigrationConfig.Enable {
				kwildCfg, genesisConfig, err = server.PrepareForMigration(cmd.Context(), kwildCfg, genesisConfig, logger)
				if err != nil {
					return fmt.Errorf("failed to prepare for migration: %w", err)
				}
			}

			logger.Info("Entering startup mode")
			if err := genesisConfig.SanityChecks(); err != nil {
				return fmt.Errorf("genesis configuration failed sanity checks: %w", err)
			}

			// Set the chain package's active forks variable. This provides easy
			// access to important chain config to other high level app packages.
			chain.SetForks(genesisConfig.ForkHeights)

			fmt.Printf("Started with %d configured hard fork heights:\n%v\n",
				len(genesisConfig.ForkHeights), genesisConfig.Forks())

			if autoGen && !configFileExists { // write config.toml if it didn't exist, and doing autogen
				cfgPath := kwildcfg.ConfigFilePath(kwildCfg.RootDir)
				fmt.Printf("Writing config file to %v\n", cfgPath)
				err = nodecfg.WriteConfigFile(cfgPath, kwildCfg)
				if err != nil {
					return err
				}
			}

			stopProfiler, err := startProfilers(kwildCfg)
			if err != nil {
				cmd.Usage()
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

			svr, err := server.New(ctx, kwildCfg, genesisConfig, nodeKey, autoGen, logger)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil // early but clean shutdown
				}
				return err
			}

			return svr.Start(ctx)
		},
	}

	flagSet := cmd1.Flags()
	flagSet.SortFlags = false
	kwildcfg.AddConfigFlags(flagSet, flagCfg)
	viper.BindPFlags(flagSet)

	flagSet.BoolVarP(&autoGen, "autogen", "a", false,
		"auto generate private key, genesis file, and config file if not exist")

	return cmd1
}

// parseExtensionFlags parses the extension flags from the command line and
// returns a map of extension names to their configured values
func parseExtensionFlags(args []string) (map[string]map[string]string, error) {
	exts := make(map[string]map[string]string)
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--extension.") {
			return nil, fmt.Errorf("expected extension flag, got %q", args[i])
		}
		// split the flag into the extension name and the flag name
		// we intentionally do not use SplitN because we want to verify
		// there are exactly 3 parts.
		parts := strings.Split(args[i], ".")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid extension flag %q", args[i])
		}

		extName := parts[1]

		// get the extension map for the extension name.
		// if it doesn't exist, create it.
		ext, ok := exts[extName]
		if !ok {
			ext = make(map[string]string)
			exts[extName] = ext
		}

		// we now need to get the flag value. Flags can be passed
		// as either "--extension.extname.flagname value" or
		// "--extension.extname.flagname=value"
		if strings.Contains(parts[2], "=") {
			// flag value is in the same argument
			val := strings.SplitN(parts[2], "=", 2)
			ext[val[0]] = val[1]
		} else {
			// flag value is in the next argument
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for extension flag %q", args[i])
			}

			if strings.HasPrefix(args[i+1], "--") {
				return nil, fmt.Errorf("missing value for extension flag %q", args[i])
			}

			ext[parts[2]] = args[i+1]
			i++
		}
	}

	return exts, nil
}

func startProfilers(cfg *config.KwildConfig) (func(), error) {
	mode := cfg.AppConfig.ProfileMode
	pprofFile := cfg.AppConfig.ProfileFile
	if pprofFile == "" {
		pprofFile = fmt.Sprintf("kwild-%s.pprof", mode)
	}

	switch cfg.AppConfig.ProfileMode {
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
		return nil, fmt.Errorf("unknown profile mode %s", cfg.AppConfig.ProfileMode)
	}
}
