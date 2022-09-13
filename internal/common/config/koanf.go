package config

import (
	"github.com/knadh/koanf"
)

type koanfConfig struct {
	*koanf.Koanf
}

func (c *koanfConfig) Select(path string) Config {
	return &koanfConfig{c.Cut(path)}
}

func (c *koanfConfig) As(out interface{}) error {
	return c.Unmarshal("", out)
}

func (c *koanfConfig) ToMap() map[string]any {
	return c.Raw()
}
