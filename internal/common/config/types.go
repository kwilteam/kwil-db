package config

type ConfigBuilder interface {
	UseEnv() ConfigBuilder
	UseFile(string) ConfigBuilder
	UseMap(map[string]string) ConfigBuilder
	WithRoot(root string)

	Build() (Config, error)
}

type Config interface {
	Select(key string) Config

	As(out interface{}) error
	Exists(key string) bool

	String(key string) string
	Int32(key string, defaultValue int32) int32
	Int64(key string, defaultValue int64) int64

	GetString(key string, defaultValue string) string
	GetInt32(key string, defaultValue int32) (int32, error)
	GetInt64(key string, defaultValue int64) (int64, error)

	ToMap() map[string]string
}
