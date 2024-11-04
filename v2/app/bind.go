package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/cobra"
)

func PreRunBindFlags(cmd *cobra.Command, args []string) error {
	// Load flags
	// mergeFn := koanf.WithMergeFunc(func(src, dest map[string]interface{}) error {
	// 	// custom merge logic, copying values from src into dst
	// 	return nil
	// })
	if err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil /*mergeFn*/); err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}
	return nil
}

func PreRunBindEnv(cmd *cobra.Command, args []string) error {
	k.Load(env.Provider("KWIL_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "KWIL_")), "_", ".", -1)
	}), nil)

	// fmt.Println(k.Sprint()) // show all merged conf
	return nil
}

// PreRunBindConfigFile loads and merges settings from the config file.
func PreRunBindConfigFile(cmd *cobra.Command, args []string) error {
	// Grab the root directory like other commands that will access it via
	// persistent flags from the root command.
	rootDir, err := RootDir(cmd)
	if err != nil {
		return err // a parent command needs to have a persistent flag named "root"
	}

	// Load config from file
	confPath, _ := filepath.Abs(filepath.Join(rootDir, "config.toml"))
	if err := k.Load(file.Provider(confPath), toml.Parser()); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error loading config: %v", err)
		}
		// Not an error, just no config file present.
		// fmt.Println("No config file present at", confPath)
	}
	return nil
}

type PreRunCmd func(*cobra.Command, []string) error

func ChainPreRuns(preRuns ...PreRunCmd) PreRunCmd {
	return func(cmd *cobra.Command, args []string) error {
		for _, preRun := range preRuns {
			if err := preRun(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

// nolint WIP
