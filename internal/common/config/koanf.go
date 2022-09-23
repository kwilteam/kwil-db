package config

import (
	"errors"
	"reflect"
	"time"

	"github.com/knadh/koanf"
)

var (
	ErrUnsupportedType = errors.New("unsupported type")
)

type koanfConfig struct {
	*koanf.Koanf
}

func (c *koanfConfig) Select(path string) Config {
	return &koanfConfig{c.Cut(path)}
}

func (c *koanfConfig) Extract(key string, v any) error {
	switch t := v.(type) {
	case *int:
		*t = c.Int(key)
	case *int8:
		*t = int8(c.Int(key))
	case *int16:
		*t = int16(c.Int(key))
	case *int32:
		*t = int32(c.Int(key))
	case *uint8:
		*t = uint8(c.Int(key))
	case *uint16:
		*t = uint16(c.Int(key))
	case *uint32:
		*t = uint32(c.Int(key))
	case *uint64:
		*t = uint64(c.Int64(key))
	case *int64:
		*t = c.Int64(key)
	case *float32:
		*t = float32(c.Float64(key))
	case *float64:
		*t = c.Float64(key)
	case *string:
		*t = c.String(key)
	case *bool:
		*t = c.Bool(key)
	case *time.Duration:
		*t = c.Duration(key)
	case *time.Time:
		*t = c.Time(key, time.RFC3339)
	case *[]int:
		*t = c.Ints(key)
	case *[]byte:
		*t = c.Bytes(key)
	case *[]int64:
		*t = c.Int64s(key)
	case *[]float64:
		*t = c.Float64s(key)
	case *[]string:
		*t = c.Strings(key)
	case *[]bool:
		*t = c.Bools(key)
	case *map[string]int:
		*t = c.IntMap(key)
	case *map[string]any:
		*t = c.Raw()
	case *map[string]int64:
		*t = c.Int64Map(key)
	case *map[string]float64:
		*t = c.Float64Map(key)
	case *map[string]string:
		*t = c.StringMap(key)
	case *map[string][]string:
		*t = c.StringsMap(key)
	case *map[string]bool:
		*t = c.BoolMap(key)
	default:
		if reflect.TypeOf(t).Kind() == reflect.Struct {
			return c.Unmarshal(key, v)
		}
		return ErrUnsupportedType
	}
	return nil
}
