package cfgx

import (
	"errors"
	"flag"
	"fmt"
	"kwil/x/utils"
	"log"
	"math"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	cfgOnce       sync.Once
	defaultConfig Config

	testCfgOnce       sync.Once
	testDefaultConfig Config
)

type config struct {
	root   map[string]string
	source map[any]any
	prefix string
}

func (c *config) Extract(key string, out any) error {
	return c.Select(key).As(out)
}

func (c *config) As(out any) error {
	b, err := yaml.Marshal(c.source)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, out)
}

func (c *config) Exists(key string) bool {
	if _, ok := c.root[c.normalize(key)]; ok {
		return true
	}
	return false
}

func (c *config) ToStringMap() map[string]string {
	result := make(map[string]string)

	for k, v := range c.root {
		if strings.HasPrefix(k, c.prefix) {
			k2 := strings.TrimPrefix(k, c.prefix)
			result[k2] = v
		}
	}

	return result
}

func (c *config) ToMap() map[string]any {
	result := make(map[string]any)

	for k, v := range c.source {
		key := k.(string)
		result[key] = v
	}

	return result
}

func (c *config) Select(key string) Config {
	if key == "" {
		return c
	}

	m, ok := c.source[key]
	if !ok {
		return emptyConfig
	}

	var k = c.normalize(key) + "."
	if ok && reflect.TypeOf(m).Kind() == reflect.Map {
		return &config{c.root, m.(map[any]any), k}
	}

	return &config{c.root, make(map[any]any), k}
}

func (c *config) String(key string) string {
	return c.GetString(key, "")
}

func (c *config) GetString(key string, defaultValue string) string {
	if v, ok := c.root[c.normalize(key)]; ok {
		return reflect.ValueOf(v).String()
	}

	return defaultValue
}

func (c *config) StringSlice(key string, delimiter string) []string {
	v := c.GetStringSlice(key, delimiter, nil)
	if v == nil {
		return []string{}
	}

	return v
}

func (c *config) GetStringSlice(key string, delimiter string, defaultValue []string) []string {
	v := c.String(key)
	if v == "" {
		return defaultValue
	}

	return strings.Split(v, delimiter)
}

func (c *config) Int32(key string, defaultValue int32) int32 {
	v, err := c.GetInt32(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as int32: %v\n", key, err)
	return defaultValue
}

func (c *config) UInt32(key string, defaultValue uint32) uint32 {
	v, err := c.GetUInt32(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as uint32: %v\n", key, err)
	return defaultValue
}

func (c *config) Int64(key string, defaultValue int64) int64 {
	v, err := c.GetInt64(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as int64: %v\n", key, err)
	return defaultValue
}

func (c *config) UInt64(key string, defaultValue uint64) uint64 {
	v, err := c.GetUInt64(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as uint64: %v\n", key, err)
	return defaultValue
}

func (c *config) GetInt32(key string, defaultValue int32) (int32, error) {
	s := c.String(key)
	if s == "" {
		return defaultValue, nil
	}

	result, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return 0, err
	}

	return int32(result), nil
}

func (c *config) GetUInt32(key string, defaultValue uint32) (uint32, error) {
	s := c.String(key)
	if s == "" {
		return defaultValue, nil
	}

	result, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return 0, err
	}

	if result > math.MaxUint32 {
		return 0, errors.New("value is too big")
	}

	return uint32(result), nil
}

func (c *config) GetInt64(key string, defaultValue int64) (int64, error) {
	s := c.String(key)
	if s == "" {
		return defaultValue, nil
	}

	result, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (c *config) GetUInt64(key string, defaultValue uint64) (uint64, error) {
	s := c.String(key)
	if s == "" {
		return defaultValue, nil
	}

	result, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (c *config) Bool(key string, defaultValue bool) bool {
	v, err := c.GetBool(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as bool: %v\n", key, err)
	return defaultValue
}

func (c *config) GetBool(key string, defaultValue bool) (bool, error) {
	s := c.String(key)
	if s == "" {
		return defaultValue, nil
	}

	return strconv.ParseBool(s)
}

func (c *config) Duration(key string, defaultValue time.Duration) time.Duration {
	v, err := c.GetDuration(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as Duration: %v\n", key, err)
	return defaultValue
}

func (c *config) GetDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	s := c.String(key)
	if s == "" {
		return defaultValue, nil
	}

	return time.ParseDuration(s)
}

func (c *config) normalize(key string) string {
	return c.prefix + key
}

var loadedConfigSources []Source
var loadedTestConfigSources []Source

func getConfigSourcesInternal() []Source {
	getConfigInternal() //ensure it is loaded

	var local []Source
	return append(local, loadedConfigSources...)
}
func getTestConfigSourcesInternal() []Source {
	getTestConfigInternal() //ensure it is loaded

	var local []Source
	return append(local, loadedTestConfigSources...)
}

func getTestConfigInternal() Config {
	return _getConfigInternal(true)
}

func getConfigInternal() Config {
	return _getConfigInternal(false)
}

func getConfigFile(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return ""
	}

	return path
}

const ENV_KEY_PREFIX_DEFAULT = "kenv"
const ENV_OS_KEY_FILTER = "env-key-filter"
const ENV_SETTINGS_PATH = "env-settings"
const Meta_Config_Flag = "meta-config"
const Meta_Config_Test_Flag = "meta-config-test"

const Root_Dir_Flag = "meta-config-root-dir"
const Root_Test_Dir_Flag = "meta-config-root-dir-test"

const Root_Dir_Env = ENV_KEY_PREFIX_DEFAULT + "_" + Root_Dir_Flag
const Root_Test_Dir_Env = ENV_KEY_PREFIX_DEFAULT + "_" + Root_Test_Dir_Flag

const Meta_Config_Env = ENV_KEY_PREFIX_DEFAULT + "_" + Meta_Config_Flag
const Meta_Config_Test_Env = ENV_KEY_PREFIX_DEFAULT + "_" + Meta_Config_Test_Flag

func _getConfigInternal(test bool) Config {
	once := utils.IfElse(test, &testCfgOnce, &cfgOnce)

	//should look for a metaConfig to specify things like useEnv, various files, etc
	once.Do(func() {
		root_dir := os.Getenv(utils.IfElse(test, Root_Test_Dir_Env, Root_Dir_Env))
		if root_dir == "" {
			root_dir = "./"
		}
		file := utils.IfElse(test, Meta_Config_Test_Flag, Meta_Config_Flag)
		configFile := *flag.String(file, "", "Path to configuration file")
		flag.Parse()
		if configFile == "" {
			configFile = os.Getenv(ENV_KEY_PREFIX_DEFAULT + "_" + file)
			if configFile == "" {
				configFile = getConfigFile(path.Join(root_dir, file) + ".yaml")
				if configFile == "" {
					configFile = getConfigFile(path.Join(root_dir, file) + ".yml")
					if configFile == "" {
						configFile = getConfigFile(path.Join(root_dir, file) + ".json")
						if configFile == "" {
							fmt.Println(getConfigFileUsage())
							os.Exit(2)
						}
					}
				}
			}
		}

		rootBuilder := &config_builder{}

		cfg, err := rootBuilder.UseFile("", configFile).Build()
		if err != nil {
			panic(err)
		}

		var filter string
		if utils.IsRunningInContainer() {
			filter = ""
		} else if envSettings := cfg.GetString(ENV_SETTINGS_PATH, ""); envSettings != "" {
			filter = cfg.GetString(ENV_OS_KEY_FILTER, ENV_KEY_PREFIX_DEFAULT)
			if !strings.HasSuffix(filter, "_") {
				filter += "_"
			}

			env, err := Builder().UseFile(ENV_SETTINGS_PATH, envSettings).Build()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Using env settings from %s\n", envSettings)
			for k, v := range env.ToStringMap() {
				if !strings.HasPrefix(k, filter) {
					k = filter + k
				}
				err := os.Setenv(k, os.ExpandEnv(v))
				if err != nil {
					panic(err)
				}
			}
		}

		b := Builder().UseEnv(filter)
		for k, v := range cfg.ToStringMap() {
			if k == ENV_SETTINGS_PATH || k == ENV_OS_KEY_FILTER {
				continue
			}

			values := strings.Split(v, ",")
			if len(values) == 1 {
				b = b.UseFile(k, path.Join(root_dir, v))
			} else if len(values) == 2 {
				b = b.UseFileSelection(k, strings.TrimSpace(values[0]), path.Join(root_dir, strings.TrimSpace(values[1])))
			} else {
				panic("invalid config value: " + v)
			}
		}

		cfg, err = b.Build()
		if err != nil {
			panic(err)
		}

		if test {
			loadedTestConfigSources = rootBuilder.sources
			testDefaultConfig = cfg
		} else {
			loadedConfigSources = rootBuilder.sources
			defaultConfig = cfg
		}
	})

	return utils.IfElse(test, testDefaultConfig, defaultConfig)
}

func getConfigFileUsage() string {
	return "No meta-config file found. " +
		"By default, the lookup logic is as follows:\n" +
		"a) Use the path specified on the command line via --" + Meta_Config_Flag + "\n" +
		"b) Use the path specified in the environment variable '" + Meta_Config_Env + "'\n" +
		"c) Look in the current application's working directory for a file called\n" +
		"   " + Meta_Config_Flag + ".yaml or " + Meta_Config_Flag + ".yml.\n\n" +
		"Inside of the resolved meta config, a section called '" + ENV_SETTINGS_PATH + "' is used to inject\n" +
		"key/value pairs into the environment variables via os.Setenv(). This is done prior\n" +
		"to parsing config files specified in the meta config. All config files resolved are then\n" +
		"parsed using os.ExpandEnv() to inject any values for placeholders specified as environment\n" +
		"variable names (subject to the same rules indicated for os.ExpandEnv())\n"
}
