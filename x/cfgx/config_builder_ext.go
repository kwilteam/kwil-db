package cfgx

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type config_builder struct {
	root    string
	sources []Source
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
			//this should not happen, but just in case
			panic("only one source is supported")
		}

		m := make(map[interface{}]interface{})
		err := source.Sources()[0].As(m)
		if err != nil {
			return nil, err
		}

		rootMap[source.Name()] = m
	}

	flattenedMap := make(map[string]string)

	expand(rootMap)

	flatten(flattenedMap, b.root, rootMap)

	return &config{flattenedMap, rootMap, b.root}, nil
}

func (b *config_builder) UseFile(name string, path string) ConfigBuilder {
	source := createConfigSource(name, createConfigFileSource(path))
	return &config_builder{b.root, append(b.sources, source)}
}

func (b *config_builder) UseFileSelection(name string, selector string, path string) ConfigBuilder {
	source := createConfigSource(name, createConfigFileSelectorSource(path, selector))
	return &config_builder{b.root, append(b.sources, source)}
}

func expand(m map[interface{}]interface{}) {
	for k, v := range m {
		kind := reflect.TypeOf(v).Kind()
		switch kind {
		case reflect.Map:
			expand(v.(map[interface{}]interface{}))
		case reflect.String:
			m[k] = os.ExpandEnv(v.(string))
		}
	}
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

//func (b *config_builder) UseEnv(filter x.Predicate[string]) ConfigBuilder {
//	if filter == nil {
//		filter = x.TruePredicate[string]()
//	}
//	return &config_builder{b.root, append(b.sources, &env_source{filter: filter})}
//}

//func (b *config_builder) UseSource(source Source) ConfigBuilder {
//	return &config_builder{b.root, append(b.sources, source)}
//}
