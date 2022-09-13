package config

import "time"

type ConfigBuilder interface {
	UseEnv(string) ConfigBuilder
	UseFile(string) ConfigBuilder
	UseMap(map[string]any) ConfigBuilder
	Build() (Config, error)
}

type Config interface {
	Select(key string) Config
	As(out interface{}) error
	ToMap() map[string]any
	Keys() []string

	Exists(key string) bool
	Get(string) any
	Int64(path string) int64
	Int64s(path string) []int64
	Int64Map(path string) map[string]int64
	Int(path string) int
	Ints(path string) []int
	IntMap(path string) map[string]int
	Float64(path string) float64
	Float64s(path string) []float64
	Float64Map(path string) map[string]float64
	Duration(path string) time.Duration
	Time(path, layout string) time.Time
	String(path string) string
	Strings(path string) []string
	StringMap(path string) map[string]string
	StringsMap(path string) map[string][]string
	Bytes(path string) []byte
	Bool(path string) bool
	Bools(path string) []bool
	BoolMap(path string) map[string]bool
}
