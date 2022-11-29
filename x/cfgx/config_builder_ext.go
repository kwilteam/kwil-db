package cfgx

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type config_builder struct {
	root       string
	sources    []Source
	env_prefix string
	maps       []map[string]any
}

func (b *config_builder) Build() (Config, error) {
	if len(b.sources) == 0 {
		return emptyConfig, nil
	}

	rootMap := make(map[interface{}]interface{})
	for _, source := range b.sources {
		if len(source.Sources()) == 0 {
			continue
		}

		if len(source.Sources()) > 1 {
			//this should not happen with current sources, but just in case
			panic("only one source is supported")
		}

		m := make(map[interface{}]interface{})
		err := source.Sources()[0].As(m)
		if err != nil {
			return nil, err
		}

		rootMap[source.Name()] = m
	}

	if len(b.maps) > 0 {
		for _, m := range b.maps {
			for k, v := range m {
				rootMap[k] = v
			}
		}
	}

	b.expand(rootMap)

	flattenedMap := make(map[string]string)

	if b.env_prefix != "" {
		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			if strings.HasPrefix(pair[0], b.env_prefix) {
				key := strings.TrimPrefix(pair[0], b.env_prefix)
				if key != "" {
					flattenedMap[key] = pair[1]
				}
			}
		}
	}

	flatten(flattenedMap, b.root, rootMap)

	return &config{flattenedMap, rootMap, b.root}, nil
}

func (b *config_builder) UseEnv(filter string) ConfigBuilder {
	if filter != "" && !strings.HasSuffix(filter, "_") {
		filter = filter + "_"
	}

	return &config_builder{b.root, b.sources, filter, b.maps}
}

func (b *config_builder) UseMap(m map[string]any) ConfigBuilder {
	if m == nil {
		return b
	}

	return &config_builder{b.root, b.sources, b.env_prefix, append(b.maps, m)}
}

func (b *config_builder) UseFile(name string, path string) ConfigBuilder {
	source := createConfigSource(name, createConfigFileSource(path))
	return &config_builder{b.root, append(b.sources, source), b.env_prefix, b.maps}
}

func (b *config_builder) UseFileSelection(name string, selector string, path string) ConfigBuilder {
	source := createConfigSource(name, createConfigFileSelectorSource(path, selector))
	return &config_builder{b.root, append(b.sources, source), b.env_prefix, b.maps}
}

func (b *config_builder) expand(m map[interface{}]interface{}) {
	for k, v := range m {
		if v == nil {
			m[k] = nil
			continue
		}
		kind := reflect.TypeOf(v).Kind()
		switch kind {
		case reflect.Map:
			b.expand(v.(map[interface{}]interface{}))
		case reflect.String:
			m[k] = os.Expand(v.(string), b.getEnv)
		}
	}
}

func (b *config_builder) getEnv(key string) string {
	return os.Getenv(b.env_prefix + key)
}

func flatten(flattened map[string]string, prefix string, m map[interface{}]interface{}) {
	if prefix != "" && !strings.HasSuffix(prefix, ".") {
		prefix = prefix + "."
	}

	for k, v := range m {
		var key = prefix + k.(string)
		if v == nil {
			flattened[key] = ""
			continue
		}

		if reflect.TypeOf(v).Kind() == reflect.Map {
			flatten(flattened, key, v.(map[interface{}]interface{}))
		} else {
			var t = reflect.TypeOf(v).Kind()
			switch t {
			case reflect.String:
				flattened[key] = v.(string)
			case reflect.Bool:
				flattened[key] = strconv.FormatBool(reflect.ValueOf(v).Bool())
			case reflect.Int32:
				flattened[key] = strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
			case reflect.Int64:
				flattened[key] = strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
			case reflect.Int:
				flattened[key] = strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
			case reflect.Uint32:
				flattened[key] = strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
			case reflect.Uint64:
				flattened[key] = strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
			case reflect.Uint:
				flattened[key] = strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
			default:
				fmt.Println("Unsupported type in config ('" + key + "'): " + t.String())
			}
		}
	}
}
