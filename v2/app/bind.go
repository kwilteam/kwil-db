package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kwil/log"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func RootDir(cmd *cobra.Command) (string, error) {
	// We have to get this from the flagset directly since PreRunBindConfigFile,
	// which is executed *before* PreRunBindFlags, needs to get the root
	// directory, and it is not yet loaded into k (our koanf instance).

	// if r := k.String(RootFlagName); r != "" {
	// 	return r, nil
	// }
	return cmd.Flags().GetString(RootFlagName)
}

func SetNodeFlags(cmd *cobra.Command) {

	fs := cmd.Flags()

	// top level (for now, may be common, node, etc.)
	fs.String("log-level", log.LevelInfo.String(), "log level")
	fs.String("log-format", string(log.FormatUnstructured), "log format")
	fs.BytesHex("privkey", nil, "private key to use for node")

	// [peer]
	fs.StringSlice("peer.bootnodes", nil, "bootnodes to connect to on startup")
	// fs.StringSlice("peer.seeds", nil, "seeds to get peer addresses from (for pex only, not persistent peers)")
	fs.String("peer.ip", "0.0.0.0", "ip to listen on for P2P connections")
	fs.Uint64("peer.port", 6600, "port to listen on for P2P connections")

	fs.Bool("peer.no-pex", false, "disable peer exchange") // default-false flag
	fs.Bool("peer.pex", true, "enable peer exchange")      // default-true bool flag to match toml where it is best
	cmd.Flag("peer.pex").Hidden = true                     // maybe remove if we keep default to enable pex
}

// PreRunCmd is the function signature used with a cobra.Command.PreRunE or
// PersistentPreRunE. Use [ChainPreRuns] to apply multiple PreRunCmds.
type PreRunCmd func(*cobra.Command, []string) error

// ChainPreRuns chains a list of PreRunCmd functions into a single PreRunCmd.
// This can be used to set config sources and precedence by chaining
// [PreRunBindEnvAllSections], [PreRunBindEnvMatching], and [PreRunBindConfigFile].
// Use [BindDefaults] to initialize with defaults from a config struct.
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

// BindDefaults binds a struct to the koanf instance. The tags specified which
// tag should be used to determine the names of the keys to bind for each field
// of the struct.
func BindDefaults(cfg interface{}, tag string) error {
	err := k.Load(structs.Provider(cfg, tag), nil)
	if err != nil {
		return fmt.Errorf("error loading config from struct: %v", err)
	}
	return nil
}

// The flags and config file load/merge currently standardizes to "_" instead of
// "-". With environment variables, where sections are delimited by "_", there
// is ambiguity as to whether to merge "KWIL_SECT_SOME_VALUE" with
// "sect.some.value" or "sect.some_value".

// PreRunBindFlags loads flags and standardizes the keys to "_" instead of "-".
// This is done to unify the config file and flags.
func PreRunBindFlags(cmd *cobra.Command, args []string) error {
	// Load flags
	flagSet := cmd.Flags()
	err := k.Load(posflag.ProviderWithFlag(flagSet, ".", nil, /* <- k if we want defaults from the flags*/
		func(f *pflag.Flag) (string, interface{}) {
			// if !f.Changed { debugf("not changed %v", f.Name) }
			key := strings.ToLower(f.Name)
			val := posflag.FlagVal(flagSet, f)

			if f.Changed {
				// special case translations
				switch key {
				case "peer.no-pex":
					newKey := "peer.pex"
					if valB, ok := val.(bool); ok {
						debugf("translating flag %s = %v => %s = %v", key, valB, newKey, !valB)
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
	if k.Bool("debug") {
		k.Set("log_level", "debug")
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
	kEnv := koanf.New(".")
	kEnv.Load(env.Provider("KWIL_", ".", func(s string) string {
		s, _ = strings.CutPrefix(s, "KWIL_")
		return strings.ToLower(s) // no sections here
	}), nil)

	// Merge into global k by mapping keys into env var format to find matches.
	// Matches will come from either a previously loaded default config struct
	// (PreRunBindConfigFile) or pflags (PreRunBindFlags). If neither sources
	// preceded env, they are loaded assuming every "_" is a section delimiter,
	// which may be incorrect for multi-word keys like KWIL_SECTION_SOME_KEY.
	envKeyMap := kEnv.All()         // flattened map - kEnv.Print()
	for key, val := range k.All() { // flattened keys+vals map
		envEquivKey := strings.NewReplacer(".", "_", "-", "_").Replace(key) // e.g. "section.sub.some-key" => "section_sub_some_key"
		if envVal, ok := envKeyMap[envEquivKey]; ok {
			debugf("Merging env var: %s (%v) <= %s (%v)", key, val,
				strings.ToUpper(envEquivKey), envVal)
			k.Set(key, kEnv.Get(envEquivKey))
			delete(envKeyMap, envEquivKey)
		}
	}
	// Set the remaining unmatched env vars into k.
	for key, val := range envKeyMap {
		debugf("Unmatched env var: %s", key)
		k.Set(strings.ReplaceAll(key, "_", "."), val)
	}
	return nil
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

// PreRunBindConfigFile loads and merges settings from the config file.
func PreRunBindConfigFile(cmd *cobra.Command, args []string) error {
	// Grab the root directory like other commands that will access it via
	// persistent flags from the root command.
	rootDir, err := RootDir(cmd)
	if err != nil {
		return err // a parent command needs to have a persistent flag named "root"
	}

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
	confPath, _ := filepath.Abs(filepath.Join(rootDir, ConfigFileName))
	if err := k.Load(file.Provider(confPath), toml.Parser() /*, mergeFn*/); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error loading config: %v", err)
		}
		// Not an error, just no config file present.
		debugf("No config file present at %v", confPath)
	}
	return nil
}

// PreRunPrintEffectiveConfig prints the effective config map if CLI debugging
// is enabled (the `--debug` flag is set), otherwise this does nothing. It may
// be specified multiple times in a PreRun chain.
func PreRunPrintEffectiveConfig(cmd *cobra.Command, args []string) error {
	debugf("merged config map:\n%s\n", lazyPrinter(func() string {
		return k.Sprint()
	}))
	return nil
}

// nolint WIP
func mergeWithKeyTransform(keyFn func(s string) string) func(src, dest map[string]interface{}) error {
	return func(src, dest map[string]interface{}) error {
		return mergeFunc(src, dest, keyFn)
	}
}

func mergeFunc(src, dest map[string]interface{}, keyFn func(s string) string) error {
	for key, srcVal := range src {
		// First transform the key to the desired format (e.g. remove underscores)
		key = keyFn(key)

		destVal, exists := dest[key]
		if !exists { // new
			dest[key] = srcVal
			continue
		}

		// attempt to merge, overwriting with srcVal or merging nested maps recursively
		srcMap, srcIsMap := srcVal.(map[string]interface{})
		destMap, destIsMap := destVal.(map[string]interface{})

		if srcIsMap != destIsMap {
			return fmt.Errorf("conflict: attempting to replace non-map with a map")
		}

		if srcIsMap { // both are maps
			return mergeFunc(srcMap, destMap, keyFn)
		}

		// Replace dest's non-map value with src's non-map value
		dest[key] = srcVal
	}
	return nil
}
