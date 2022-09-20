package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"go.uber.org/multierr"
)

func Builder() ConfigBuilder {
	return &configBuilderImpl{cfg: koanf.New(".")}
}

type configBuilderImpl struct {
	cfg    *koanf.Koanf
	useEnv bool
	errs   []error
}

func (b *configBuilderImpl) Build() (Config, error) {
	return &koanfConfig{Koanf: b.cfg}, multierr.Combine(b.errs...)
}

func (b *configBuilderImpl) UseEnv(prefix string) ConfigBuilder {
	err := b.cfg.Load(env.Provider(prefix, ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, prefix)), "_", ".")
	}), nil)
	if err != nil {
		b.errs = append(b.errs, err)
	}
	return b
}

func (b *configBuilderImpl) UseFile(path string) ConfigBuilder {
	var parser koanf.Parser
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		parser = yaml.Parser()
	case ".json":
		parser = json.Parser()
	case ".toml":
		parser = toml.Parser()
	default:
		b.errs = append(b.errs, fmt.Errorf("unknown file extension: %s", path))
		return b
	}

	if err := b.cfg.Load(file.Provider(path), parser); err != nil {
		b.errs = append(b.errs, err)
	}
	return b
}

func (b *configBuilderImpl) UseMap(m map[string]any) ConfigBuilder {
	if err := b.cfg.Load(confmap.Provider(m, "."), nil); err != nil {
		b.errs = append(b.errs, err)
	}
	return b
}
