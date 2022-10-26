package cfgx

type Source interface {
	Name() string
	Sources() []SourceItem
}

type SourceItem interface {
	As(out any) error
}

type FileSource interface {
	Path() string
	As(out any) error
}

type FileSelectorSource interface {
	Selector() string
	As(out any) error
}

type ConfigSource interface {
	Name() string

	// Sources currently only one source item supported
	Sources() []SourceItem
}

func GetConfigSources() []Source {
	return getConfigSourcesInternal()
}

func GetTestConfigSources() []Source {
	return getTestConfigSourcesInternal()
}
