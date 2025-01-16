// Package conf is used to merge multiple node config sources. It encapsulates
// the koanf package, which is functionally similar to viper.
//
// A cobra command should:
//
//  1. BindDefaults with a tagged struct instance with the default values set.
//  2. Use one or more "PreRun" functions, such as PreRunBindFlags,
//     PreRunBindEnvMatching, and PreRunBindConfigFile, or those in the bind package.
//  3. Use the bind.ChainPreRuns helper to combine the preruns in the desired order,
//     which dictates the merge priority for the config sources (e.g. env > flags > toml).
//  4. Use ActiveConfig to get the merged config. Any subcommand can access the
//     config this way.
//
// See also the app/shared/bind package for shared helper functions for working
// with flags and koanf instances.
package conf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/config"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The functions in this file are centered around this global koanf instance
// that combines multiple config sources according to the enabled preruns.
var k = koanf.New(".")

const koanfTag = "toml"

// ActiveConfig retrieves the current merged config. This is influenced by the
// other functions in this package, including: BindDefaults,
// SetFlagsFromStruct, PreRunBindFlags, PreRunBindEnvMatching,
// PreRunBindEnvAllSections, and PreRunBindConfigFile. The merge result is
// determined by the order in which they source are bound (see PreRunBindConfigFile).
func ActiveConfig() *config.Config {
	// k => config.Config
	var cfg config.Config
	err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: koanfTag})
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}
	// rootDir := k.String(bind.RootFlagName)
	return &cfg
}

// BindDefaults binds a struct to the koanf instance. The field names should have
// `koanf:"name"` tags to bind the correct name.
func BindDefaults(cfg any) error {
	return bind.BindDefaultsTo(cfg, koanfTag, k)
}

func BindDefaultsWithRootDir(cfg any, rootDir string) error {
	if err := bind.BindDefaultsTo(cfg, koanfTag, k); err != nil {
		return err
	}
	return k.Set(bind.RootFlagName, rootDir)
	// return k.Load(&mapProvider{map[string]any{bind.RootFlagName: rootDir}}, nil)
}

func RootDir() string {
	return k.String(bind.RootFlagName)
}

// The flags and config file load/merge currently standardizes to "_" instead of
// "-". With environment variables, where sections are delimited by "_", there
// is ambiguity as to whether to merge "KWIL_SECT_SOME_VALUE" with
// "sect.some.value" or "sect.some_value".

// PreRunBindFlags loads flags and standardizes the keys to "_" instead of "-".
// This is done to unify the config file and flags. Be sure to define the flags
// in the command's flag set. See [bind.SetFlagsFromStruct] to automate defining
// the flags from a default config struct.
func PreRunBindFlags(cmd *cobra.Command, args []string) error {
	// Load posix flags (posflag provider).
	flagSet := cmd.Flags()
	err := k.Load(posflag.ProviderWithFlag(flagSet, ".", nil, /* <- k if we want defaults from the flags*/
		func(f *pflag.Flag) (string, interface{}) {
			// if !f.Changed { Debugf("not changed %v", f.Name) }
			key := strings.ToLower(f.Name)
			val := posflag.FlagVal(flagSet, f)

			if f.Changed {
				// special case translations
				switch key {
				case "p2p.no-pex":
					newKey := "p2p.pex"
					if valB, ok := val.(bool); ok {
						bind.Debugf("translating flag %s = %v => %s = %v", key, valB, newKey, !valB)
						val = !valB // negate
						key = newKey
					}
				}
			}

			return strings.ReplaceAll(key, "-", "_"), val
		}), nil /*no parser for flags*/ /*, mergeFn*/)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}
	return nil
}

// PreRunBindEnvMatching is to accomplish all of the following:
//
//	KWIL_SECTION_SUBSECTION_SOME_KEY => section.subsection.some-key
//	KWIL_SECTION_SUBSECTION_KEYNAME => section.subsection.keyname
//	KWIL_SECTION_SOME_KEY => section.some-key
//	KWIL_SECTION_KEYNAME => section.keyname
//
// For this to work correctly, a previously loaded default config struct
// (PreRunBindConfigFile) or pflags (PreRunBindFlags). If neither sources
// preceded env, they are loaded assuming every "_" is a section delimiter,
// which may be incorrect for multi-word keys like KWIL_SECTION_SOME_KEY.
func PreRunBindEnvMatching(cmd *cobra.Command, args []string) error {
	return bind.PreRunBindEnvMatchingTo(cmd, args, "KWIL_", k)
}

// PreRunBindEnvAllSections treats all underscores as section delimiters. With
// this approach, the config merge process can work with the following
// conventions:
//
//	toml:      section.sub.two_words  (replacing "_" with "")
//	flag:    --section.sub.two-words  (replacing "-" with "")
//	env:  KWIL_SECTION_SUB_TWOWORDS   (no underscores allowed in key name!)
//
// To merge all, k.Load from each source should merge by standardizing the key
// names into "twowords", AND the `koanf` tag should match.
func PreRunBindEnvAllSections(cmd *cobra.Command, args []string) error {
	k.Load(env.Provider("KWIL_", ".", func(s string) string {
		// The following , not the above goal.
		// KWIL_SECTION_SUBSECTION_SOMEVALUE => section.subsection.somevalue
		// Values cannot have underscores in this convention!
		s, _ = strings.CutPrefix(s, "KWIL_")
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil)
	return nil
}

// PreRunBindConfigFile loads and merges settings from the config file. This
// used shared.RootDir to get the root directory. As such, the command should be
// use shared.BindRootDirVar.
func PreRunBindConfigFile(cmd *cobra.Command, args []string) error {
	// Grab the root directory like other commands that will access it via
	// persistent flags from the root command.
	rootDir, err := bind.RootDir(cmd)
	if err != nil {
		return err // a parent command needs to have a persistent flag named "root"
	}
	// rootDir := RootDir() // from k, requires BindDefaultsWithRootDir or a root field of the config struct

	// If we want to instead have space placeholders removed (to match
	// PreRunBindEnvAllSections) rather than having them be standardized to
	// underscores to match toml, use this:
	//
	// mergeFn := koanf.WithMergeFunc(mergeWithKeyTransform(func(key string) string {
	// 	return strings.ReplaceAll(strings.ToLower(key), "_", "")
	// }))
	//
	// The above can be modified to standardize to "-".

	// Load config from file
	confPath, _ := filepath.Abs(config.ConfigFilePath(rootDir))
	if err := k.Load(file.Provider(confPath), toml.Parser() /*, mergeFn*/); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error loading config: %v", err)
		}
		// Not an error, just no config file present.
		bind.Debugf("No config file present at %v", confPath)
	}
	return nil
}

// PreRunPrintEffectiveConfig prints the effective config map if CLI debugging
// is enabled (the `--debug` flag is set), otherwise this does nothing. It may
// be specified multiple times in a PreRun chain.
func PreRunPrintEffectiveConfig(cmd *cobra.Command, args []string) error {
	bind.Debugf("merged config map:\n%s\n", bind.LazyPrinter(func() string {
		return k.Sprint()
	}))
	return nil
}
