package shared

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node/types"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	gotoml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The functions in this file are centered around this global koanf instance
// that combines multiple config sources according to the enabled preruns.
var k = koanf.New(".")

// DefaultConfig is the function to return the default to all commands in this
// app package. This is a var so that a custom binary may override the defaults,
// which many commands obtain with this function.
var DefaultConfig = config.DefaultConfig

// ActiveConfig retrieves the current merged config. This is influenced by the
// other functions in this package, including: BindDefaults,
// SetFlagsFromStruct, PreRunBindFlags, PreRunBindEnvMatching,
// PreRunBindEnvAllSections, and PreRunBindConfigFile. The merge result is
// determined by the order in which they source are bound (see PreRunBindConfigFile).
func ActiveConfig() *config.Config {
	// k => config.Config
	var cfg config.Config
	err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"})
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}
	return &cfg
}

// ConfigToTOML marshals the config to TOML.
func ConfigToTOML(cfg *config.Config) (string, error) {
	rawToml, err := gotoml.Marshal(&cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to toml: %w", err)
	}
	return string(rawToml), nil
}

const (
	RootFlagName  = "root"
	rootShortName = "r"
)

// BindRootDirVar is used to bind the RootFlagName with a local command
// variable. The flag "root" does not have config file analog, so binds with
// local var, which is then available to all subcommands via RootDir(cmd).
func BindRootDirVar(cmd *cobra.Command, rootDir *string, defaultVal, desc string) {
	cmd.PersistentFlags().StringVarP(rootDir, RootFlagName, rootShortName,
		defaultVal, desc)
}

// RootDir complements BindRootDirVar by returning the value of the root
// directory flag bound by BindRootDirVar. This is available for commands that
// do not have direct access to the local string pointer bound to the flag, such
// as subcommands.
func RootDir(cmd *cobra.Command) (string, error) {
	// We have to get this from the flagset directly since PreRunBindConfigFile,
	// which is executed *before* PreRunBindFlags, needs to get the root
	// directory, and it is not yet loaded into k (our koanf instance).

	// if r := k.String(RootFlagName); r != "" {
	// 	return r, nil
	// }
	return cmd.Flags().GetString(RootFlagName)
}

// SetFlagsFromStruct is used to automate the creation of flags in the
// provided pflag set from a tagged struct. This uses the field tags:
// "toml" for the flag name, and "comment" for the help string. Any "_"
// characters in the "toml" tag are converted to "-" when defining the flag.
// This recurses into nested structs.
func SetFlagsFromStruct(fs *pflag.FlagSet, cfg interface{}) {
	SetFlagsFromStructTags(fs, cfg, "toml", "comment")
}

// SetFlagsFromStructTags is a more generalized version of [SetFlagsFromStruct]
// that allows setting the field tags used to get flag name and description.
func SetFlagsFromStructTags(fs *pflag.FlagSet, cfg interface{}, nameTag, descTag string) {
	fs.SortFlags = false

	val := reflect.ValueOf(cfg)
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	var setFlag func(field reflect.StructField, fieldVal reflect.Value, prefix string)
	setFlag = func(field reflect.StructField, fieldVal reflect.Value, prefix string) {
		// Get flag name from toml tag
		flagName := field.Tag.Get(nameTag)
		if flagName == "" {
			flagName = strings.ToLower(field.Name)
		}
		flagName, _, _ = strings.Cut(flagName, ",")
		flagName = strings.ReplaceAll(flagName, "_", "-")
		if prefix != "" {
			flagName = prefix + "." + flagName
		}

		// Get description from comment tag
		desc := field.Tag.Get(descTag)
		if desc == "" {
			desc = flagName // fallback to name if no comment
		}

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct {
			for i := range field.Type.NumField() {
				setFlag(field.Type.Field(i), fieldVal.Field(i), flagName)
			}
			return
		}

		// first catch special types like log.Level and log.Format
		switch vt := fieldVal.Interface().(type) {
		case log.Level:
			fs.String(flagName, vt.String(), desc)
			return
		case log.Format:
			fs.String(flagName, string(vt), desc)
			return
		case time.Duration:
			fs.Duration(flagName, vt, desc)
			return
		case config.Duration:
			fs.Duration(flagName, time.Duration(vt), desc)
			return
		case types.HexBytes:
			fs.BytesHex(flagName, vt, desc)
			return
		}
		// fallback to default flag set

		// Set flag based on field type
		switch field.Type.Kind() {
		case reflect.String:
			defaultVal := fieldVal.String()
			fs.String(flagName, defaultVal, desc)
		case reflect.Bool:
			defaultVal := fieldVal.Bool()
			fs.Bool(flagName, defaultVal, desc)
		case reflect.Int, reflect.Int64, reflect.Int16, reflect.Int32, reflect.Int8:
			defaultVal := fieldVal.Int()
			fs.Int64(flagName, defaultVal, desc)
		case reflect.Uint, reflect.Uint64, reflect.Uint16, reflect.Uint32, reflect.Uint8:
			defaultVal := fieldVal.Uint()
			fs.Uint64(flagName, defaultVal, desc)
		case reflect.Float64, reflect.Float32:
			defaultVal := fieldVal.Float()
			fs.Float64(flagName, defaultVal, desc)
		case reflect.Slice:
			switch sv := fieldVal.Interface().(type) {
			case []string:
				fs.StringSlice(flagName, sv, desc)

			// TODO: this will take some maint with different slice types
			// fs.IntSlice(), etc
			default:
				fmt.Printf("Unsupported slice type for flag: %T\n", sv)
			}
		}
	}

	// Process all fields
	for i := range typ.NumField() {
		setFlag(typ.Field(i), val.Field(i), "")
	}
}

/*func SetNodeFlagsFromDefaultInKoanf(cmd *cobra.Command, k *koanf.Koanf) error {
	var cfg config.Config
	err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"})
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	SetFlagsFromStruct(cmd.Flags(), &cfg)
	return nil
}*/

/* SetNodeFlags is superseded by SetFlagsFromStruct since the config struct
field tags and types provide everything needed to bind pflags without
maintaining and explicit list of flags as this did.

func SetNodeFlags(cmd *cobra.Command) {
	// TODO: do this automatically based on the node.Config struct, and the
	// tags, including `toml` and `comment`. This will probably involve
	// reflection unless the koanf instance that loaded the DefaultConfig
	// somehow can provide everything to us. For the default values given to the
	// flags, they can simply be the zero value for whatever type since we are
	// currently binding defaults via the struct from DefaultConfig, meaning the
	// flag defaults will never be used.

	fs := cmd.Flags()
	fs.SortFlags = false

	// top level (for now, may be common, node, etc.)
	fs.String("log-level", log.LevelInfo.String(), "log level")
	fs.String("log-format", string(log.FormatUnstructured), "log format")
	fs.BytesHex("privkey", nil, "private key to use for node")

	// [p2p]
	fs.StringSlice("p2p.bootnodes", nil, "bootnodes to connect to on startup")
	// fs.StringSlice("p2p.seeds", nil, "seeds to get peer addresses from (for pex only, not persistent peers)")
	fs.String("p2p.ip", "0.0.0.0", "ip to listen on for P2P connections")
	fs.Uint64("p2p.port", 6600, "port to listen on for P2P connections")

	fs.Bool("p2p.no-pex", false, "disable peer exchange") // default-false flag
	fs.Bool("p2p.pex", true, "enable peer exchange")      // default-true bool flag to match toml where it is best
	cmd.Flag("p2p.pex").Hidden = true                     // maybe remove if we keep default to enable pex

	// [consensus]
	fs.Duration("consensus.propose-timeout", 1000*time.Millisecond, "timeout for proposing a block (leader only)")
	fs.Uint64("consensus.max-block-size", 50_000_000, "maximum size of a block (in bytes)")
	fs.Uint64("consensus.max-txs-per-block", 20_000, "maximum number of transactions per block")

	...
}*/

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
	return BindDefaultsTo(cfg, tag, k)
}

// BindDefaultsTo loads the defaults from the provided config struct into a
// Koanf instance, using field names according to the given struct tag.
func BindDefaultsTo(cfg interface{}, tag string, k *koanf.Koanf) error {
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
			// if !f.Changed { Debugf("not changed %v", f.Name) }
			key := strings.ToLower(f.Name)
			val := posflag.FlagVal(flagSet, f)

			if f.Changed {
				// special case translations
				switch key {
				case "p2p.no-pex":
					newKey := "p2p.pex"
					if valB, ok := val.(bool); ok {
						Debugf("translating flag %s = %v => %s = %v", key, valB, newKey, !valB)
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
	return PreRunBindEnvMatchingTo(cmd, args, "KWIL_", k)
}

func PreRunBindEnvMatchingTo(cmd *cobra.Command, args []string, prefix string, k *koanf.Koanf) error {
	kEnv := koanf.New(".")
	kEnv.Load(env.Provider(prefix, ".", func(s string) string {
		s, _ = strings.CutPrefix(s, prefix)
		return strings.ToLower(s) // no sections here
	}), nil)

	// Merge into k by mapping keys into env var format to find matches.
	// Matches will come from either a previously loaded default config struct
	// (PreRunBindConfigFile) or pflags (PreRunBindFlags). If neither sources
	// preceded env, they are loaded assuming every "_" is a section delimiter,
	// which may be incorrect for multi-word keys like KWIL_SECTION_SOME_KEY.
	envKeyMap := kEnv.All()         // flattened map - kEnv.Print()
	for key, val := range k.All() { // flattened keys+vals map
		envEquivKey := strings.NewReplacer(".", "_", "-", "_").Replace(key) // e.g. "section.sub.some-key" => "section_sub_some_key"
		if envVal, ok := envKeyMap[envEquivKey]; ok {
			Debugf("Merging env var: %s (%v) <= %s (%v)", key, val,
				strings.ToUpper(envEquivKey), envVal)
			k.Set(key, kEnv.Get(envEquivKey))
			delete(envKeyMap, envEquivKey)
		}
	}
	// Set the remaining unmatched env vars into k.
	for key, val := range envKeyMap {
		Debugf("Unmatched env var: %s", key)
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
	confPath, _ := filepath.Abs(filepath.Join(rootDir, config.ConfigFileName))
	if err := k.Load(file.Provider(confPath), toml.Parser() /*, mergeFn*/); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error loading config: %v", err)
		}
		// Not an error, just no config file present.
		Debugf("No config file present at %v", confPath)
	}
	return nil
}

// PreRunPrintEffectiveConfig prints the effective config map if CLI debugging
// is enabled (the `--debug` flag is set), otherwise this does nothing. It may
// be specified multiple times in a PreRun chain.
func PreRunPrintEffectiveConfig(cmd *cobra.Command, args []string) error {
	Debugf("merged config map:\n%s\n", LazyPrinter(func() string {
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
			return errors.New("conflict: attempting to replace non-map with a map")
		}

		if srcIsMap { // both are maps
			return mergeFunc(srcMap, destMap, keyFn)
		}

		// Replace dest's non-map value with src's non-map value
		dest[key] = srcVal
	}
	return nil
}
