package cfgx

import "time"

type Config interface {
	Select(key string) Config

	// As will unmarshal the config into a given struct
	As(out any) error
	Exists(key string) bool

	Extract(key string, out any) error

	String(key string) string
	GetString(key string, defaultValue string) string

	StringSlice(key string, delimiter string) []string
	GetStringSlice(key string, delimiter string, defaultValue []string) []string

	Bool(key string, defaultValue bool) bool
	GetBool(key string, defaultValue bool) (bool, error)

	Duration(key string, defaultValue time.Duration) time.Duration
	GetDuration(key string, defaultValue time.Duration) (time.Duration, error)

	Int32(key string, defaultValue int32) int32
	GetInt32(key string, defaultValue int32) (int32, error)

	Int64(key string, defaultValue int64) int64
	GetInt64(key string, defaultValue int64) (int64, error)

	UInt32(key string, defaultValue uint32) uint32
	GetUInt32(key string, defaultValue uint32) (uint32, error)

	UInt64(key string, defaultValue uint64) uint64
	GetUInt64(key string, defaultValue uint64) (uint64, error)

	ToStringMap() map[string]string
	ToMap() map[string]any
}
