package config

func GetConfigSources() []ConfigSource {
	return getConfigSourcesInternal()
}

func GetConfig() Config {
	return getConfigInteral()
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

	ToStringMap() map[string]string
	ToMap() map[string]interface{}
}

type ConfigSource interface {
	Name() string
	Sources() []ConfigSourceItem
}

type ConfigSourceItem interface {
	As(out interface{}) error
}

type ConfigFileSource interface {
	ConfigSourceItem
	Path() string
}

type ConfigFileSelectorSource interface {
	ConfigFileSource
	Selector() string
}
