package cfgx

type ConfigBuilder interface {
	UseFile(name string, path string) ConfigBuilder
	UseFileSelection(name string, selector string, path string) ConfigBuilder

	Build() (Config, error)
}

func Builder() ConfigBuilder {
	return &config_builder{}
}
