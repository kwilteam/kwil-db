package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

func Builder() ConfigBuilder {
	return &configBuilderImpl{}
}

type configBuilderImpl struct {
	root      string
	configFns []func() (map[interface{}]interface{}, error)
}

func (b *configBuilderImpl) Build() (Config, error) {
	if len(b.configFns) == 0 {
		return emptyConfig, nil
	}

	combined := make(map[interface{}]interface{})

	for _, fn := range b.configFns {
		m, err := fn()
		if err != nil {
			return nil, err
		}

		for k, v := range m {
			combined[k] = v
		}
	}

	result := make(map[string]string)
	flatten(result, b.root, combined)

	return &configImpl{result, combined, b.root}, nil
}

func (b *configBuilderImpl) WithRoot(root string) {
	if root != "" {
		root = root + "."
	}
	b.root = root
}

func (b *configBuilderImpl) UseEnv() ConfigBuilder {
	return &configBuilderImpl{b.root, append(b.configFns, func() (map[interface{}]interface{}, error) {
		m := make(map[interface{}]interface{})
		for _, item := range os.Environ() {
			splits := strings.Split(item, "=")
			m[splits[0]] = splits[1]
		}

		return m, nil
	})}
}

func (b *configBuilderImpl) UseFile(path string) ConfigBuilder {
	return &configBuilderImpl{b.root, append(b.configFns, func() (map[interface{}]interface{}, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		return loadFromRaw(data)
	})}
}

func (b *configBuilderImpl) UseMap(m map[string]string) ConfigBuilder {
	return &configBuilderImpl{b.root, append(b.configFns, func() (map[interface{}]interface{}, error) {
		copy := make(map[interface{}]interface{})

		for k, v := range m {
			copy[k] = v
		}

		return copy, nil
	})}
}

func loadFromRaw(data []byte) (map[interface{}]interface{}, error) {
	m := make(map[interface{}]interface{})

	err := yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func flatten(flattened map[string]string, prefix string, m map[interface{}]interface{}) {
	if prefix != "" && !strings.HasSuffix(prefix, ".") {
		prefix = prefix + "."
	}

	for k, v := range m {
		var key = prefix + k.(string)
		if reflect.TypeOf(v).Kind() == reflect.Map {
			flatten(flattened, key, v.(map[interface{}]interface{}))
		} else {
			var t = reflect.TypeOf(v).Kind()
			if t == reflect.String {
				flattened[key] = v.(string)
			} else if t == reflect.Int || t == reflect.Uint {
				flattened[key] = strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
			} else if t == reflect.Uint {
				flattened[key] = strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
			} else if t == reflect.Bool {
				flattened[key] = strconv.FormatBool(reflect.ValueOf(v).Bool())
			} else {
				panic("Unsupported type in config ('" + key + "'): " + t.String())
			}
		}
	}
}
