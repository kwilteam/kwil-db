package config

type ConfigBuilder interface {
	UseEnv(string) ConfigBuilder
	UseFile(string) ConfigBuilder
	UseMap(map[string]any) ConfigBuilder
	Build() (Config, error)
}

type Config interface {
	Select(key string) Config
	Keys() []string
	Exists(key string) bool
	Get(string) any
	Extract(string, any) error

	String(string) string
	Int(string) int
}
