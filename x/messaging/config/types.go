package config

func GetConfigSources() []Source {
	return getConfigSourcesInternal()
}

func GetTestConfigSources() []Source {
	return getTestConfigSourcesInternal()
}

func GetConfig() Config {
	return getConfigInternal()
}

func GetTestConfig() Config {
	return getTestConfigInternal()
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

type Source interface {
	Name() string
	Sources() []SourceItem
}

type SourceItem interface {
	As(out interface{}) error
}

type FileSource interface {
	SourceItem
	Path() string
}

type FileSelectorSource interface {
	SourceItem
	Selector() string
}
