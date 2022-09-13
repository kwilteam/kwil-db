package config

import (
	"errors"
	"flag"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	cfgOnce       sync.Once
	defaultConfig Config
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

func (c *configImpl) ToMap() map[string]string {
	result := make(map[string]string)

	for k, v := range c.root {
		if strings.HasPrefix(k, c.prefix) {
			k2 := strings.TrimPrefix(k, c.prefix)
			result[k2] = v
		}
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
		s := reflect.ValueOf(v).String()
		return s
	}

	return defaultValue
}

func (c *configImpl) Int32(key string, defaultValue int32) int32 {
	v, err := c.GetInt32(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Println("Failed to parse num as int32", err)
	return defaultValue
}

func (c *configImpl) Int64(key string, defaultValue int64) int64 {
	v, err := c.GetInt64(key, defaultValue)
	if err == nil {
		return v
	}

	log.Default().Println("Failed to parse num as int64", err)
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

func (c *configImpl) GetInt64(key string, defaultValue int64) (int64, error) {
	s := c.String(key)
	if s == "" {
		return defaultValue, nil
	}

	result, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return 0, err
	}

	return int64(result), nil
}

func (c *configImpl) normalize(key string) string {
	return c.prefix + key
}

func GetConfig() Config {
	// should look for a metaConfig.etc to specify things like useEnv, various files, etc
	cfgOnce.Do(func() {
		var configFile string
		flag.StringVar(&configFile, "cfg", "", "Path to configuration file")

		useNoEnv := flag.Bool("cfg-no-env", false, "Do NOT use  environment variables")
		flag.Parse()
		if configFile == "" {
			// look for yaml config (e.g., config.yaml, config.yml, app-config.yaml, etc)
			var prefix = getAppNameOrDefault("app")
			configFile = getConfigFile("./" + prefix + "-config.yaml")
			if configFile == "" {
				configFile = getConfigFile("./" + prefix + "-config.yml")
				if configFile == "" {
					configFile = getConfigFile("./" + prefix + "-config.json")
					if configFile == "" {
						flag.Usage()
						os.Exit(2)
					}
				}
			}
		}

		b := Builder()
		if !*useNoEnv {
			b = b.UseEnv()
		}

		cfg, err := b.UseFile(configFile).Build()
		if err != nil {
			panic(err)
		}

		defaultConfig = cfg
	})

	return defaultConfig
}

func getConfigFile(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return ""
	}

	return path
}

func getAppNameOrDefault(defaultName string) string {
	appName, err := os.Executable()
	if err != nil {
		return defaultName
	}

	if strings.HasSuffix(appName, "/__debug_bin") {
		return defaultName
	}

	if strings.HasPrefix(appName, "/tmp/GoLand") {
		return defaultName
	}

	return appName
}
