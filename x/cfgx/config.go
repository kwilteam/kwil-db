package cfgx

import (
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"kwil/x/utils"
	"log"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	cfgOnce       sync.Once
	defaultConfig Config

	testCfgOnce       sync.Once
	testDefaultConfig Config
)

type configImpl struct {
	root   map[string]string
	source map[interface{}]interface{}
	prefix string
}

func (c *configImpl) As(out interface{}) error {
	b, err := yaml.Marshal(c.source)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, out)
}

func (c *configImpl) Exists(key string) bool {
	if _, ok := c.root[c.normalize(key)]; ok {
		return true
	}
	return false
}

func (c *configImpl) ToStringMap() map[string]string {
	result := make(map[string]string)

	for k, v := range c.root {
		if strings.HasPrefix(k, c.prefix) {
			k2 := strings.TrimPrefix(k, c.prefix)
			result[k2] = v
		}
	}

	return result
}

func (c *configImpl) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range c.source {
		key := k.(string)
		result[key] = v
	}

	return result
}

func (c *configImpl) Select(key string) Config {
	if key == "" {
		return c
	}

	m, ok := c.source[key]
	if !ok {
		return emptyConfig
	}

	var k = c.normalize(key) + "."
	if ok && reflect.TypeOf(m).Kind() == reflect.Map {
		return &configImpl{c.root, m.(map[interface{}]interface{}), k}
	}

	return &configImpl{c.root, make(map[interface{}]interface{}), k}
}

func (c *configImpl) String(key string) string {
	return c.GetString(key, "")
}

func (c *configImpl) GetString(key string, defaultValue string) string {
	if v, ok := c.root[c.normalize(key)]; ok {
		return reflect.ValueOf(v).String()
	}

	return defaultValue
}

func (c *configImpl) StringSlice(key string, delimiter string) []string {
	v := c.GetStringSlice(key, delimiter, nil)
	if v == nil {
		return []string{}
	}

	return v
}

func (c *configImpl) GetStringSlice(key string, delimiter string, defaultValue []string) []string {
	v := c.String(key)
	if v == "" {
		return defaultValue
	}

	return strings.Split(v, delimiter)
}

func (c *configImpl) Int32(key string, defaultValue int32) int32 {
	v, err := c.GetInt32(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as int32: %v\n", key, err)
	return defaultValue
}

func (c *configImpl) UInt32(key string, defaultValue uint32) uint32 {
	v, err := c.GetUInt32(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as uint32: %v\n", key, err)
	return defaultValue
}

func (c *configImpl) Int64(key string, defaultValue int64) int64 {
	v, err := c.GetInt64(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as int64: %v\n", key, err)
	return defaultValue
}

func (c *configImpl) UInt64(key string, defaultValue uint64) uint64 {
	v, err := c.GetUInt64(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as uint64: %v\n", key, err)
	return defaultValue
}

func (c *configImpl) GetInt32(key string, defaultValue int32) (int32, error) {
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

func (c *configImpl) GetUInt32(key string, defaultValue uint32) (uint32, error) {
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

func (c *configImpl) GetInt64(key string, defaultValue int64) (int64, error) {
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

func (c *configImpl) GetUInt64(key string, defaultValue uint64) (uint64, error) {
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

func (c *configImpl) Bool(key string, defaultValue bool) bool {
	v, err := c.GetBool(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as bool: %v\n", key, err)
	return defaultValue
}

func (c *configImpl) GetBool(key string, defaultValue bool) (bool, error) {
	s := c.String(key)
	if s == "" {
		return defaultValue, nil
	}

	return strconv.ParseBool(s)
}

func (c *configImpl) Duration(key string, defaultValue time.Duration) time.Duration {
	v, err := c.GetDuration(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Printf("Failed to parse %s as Duration: %v\n", key, err)
	return defaultValue
}

func (c *configImpl) GetDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	s := c.String(key)
	if s == "" {
		return defaultValue, nil
	}

	return time.ParseDuration(s)
}

func (c *configImpl) normalize(key string) string {
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

const ENV_SETTINGS_PATH = "env-settings"
const Meta_Config_Flag = "meta-config"
const Meta_Config_Test_Flag = "meta-config-test"

const Meta_Config_Env = "kenv." + Meta_Config_Flag
const Meta_Config_Test_Env = "kenv." + Meta_Config_Test_Flag

func _getConfigInternal(test bool) Config {
	once := utils.IfElse(test, &testCfgOnce, &cfgOnce)

	//should look for a metaConfig to specify things like useEnv, various files, etc
	once.Do(func() {
		file := utils.IfElse(test, Meta_Config_Test_Flag, Meta_Config_Flag)
		configFile := *flag.String(file, "", "Path to configuration file")
		flag.Parse()
		if configFile == "" {
			configFile = os.Getenv("kenv." + file)
			if configFile == "" {
				configFile = getConfigFile("./" + file + ".yaml")
				if configFile == "" {
					configFile = getConfigFile("./" + file + ".yml")
					if configFile == "" {
						configFile = getConfigFile("./" + file + ".json")
						if configFile == "" {
							fmt.Println(getConfigFileUsage())
							os.Exit(2)
						}
					}
				}
			}
		}

		rootBuilder := &configBuilderImpl{}

		cfg, err := rootBuilder.UseFile("", configFile).Build()
		if err != nil {
			panic(err)
		}

		if envSettings := cfg.GetString("env-settings", ""); envSettings != "" {
			env, err := builder().UseFile("kenv", envSettings).Build()
			if err != nil {
				panic(err)
			}

			for k, v := range env.ToStringMap() {
				err := os.Setenv(k, os.ExpandEnv(v))
				if err != nil {
					panic(err)
				}
			}
		}

		b := builder()
		for k, v := range cfg.ToStringMap() {
			if k == "env-settings" {
				continue
			}
			values := strings.Split(v, ",")
			if len(values) == 1 {
				b = b.UseFile(k, v)
			} else if len(values) == 2 {
				b = b.UseFileSelection(k, strings.TrimSpace(values[0]), strings.TrimSpace(values[1]))
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
