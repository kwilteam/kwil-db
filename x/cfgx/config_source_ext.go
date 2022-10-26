package cfgx

import (
	"fmt"
	"os"
	"path/filepath"

	"kwil/x/utils"

	"gopkg.in/yaml.v2"
)

type config_source struct {
	name    string
	sources []SourceItem
}

func (c *config_source) Sources() []SourceItem {
	return c.sources
}

func createConfigSource(name string, sources ...SourceItem) Source {
	return &config_source{name: name, sources: sources}
}

func createConfigFileSource(path string) FileSource {
	return &config_file_source{path: path}
}

func createConfigFileSelectorSource(path string, selector string) FileSelectorSource {
	return &config_file_selector_source{path: path, selector: selector}
}

func (c *config_source) add(source SourceItem) {
	c.sources = append(c.sources, source)
}

func (c *config_source) Name() string {
	return c.name
}

func (c *config_source) Items() []SourceItem {
	var local []SourceItem
	copy(local, c.sources)
	return local
}

type config_file_source struct {
	path string
}

func (c *config_file_source) Path() string {
	return c.path
}

type config_file_selector_source struct {
	path     string
	selector string
}

func (c *config_file_selector_source) Path() string {
	return c.path
}

func (c *config_file_selector_source) Selector() string {
	return c.selector
}

func (c *config_file_source) As(out any) error {
	return loadAs(c.path, out)
}

func (c *config_file_selector_source) As(out any) error {
	m := make(map[any]any)
	err := loadAs(c.path, m)
	if err != nil {
		return err
	}

	selected, ok := m[c.selector]
	if !ok {
		return fmt.Errorf("selector path (%s()) not found for source  (%s)", c.selector, c.path)
	}

	b, err := yaml.Marshal(selected)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, out)
}

func loadAs(path string, out any) error {
	switch filepath.Ext(path) {
	case ".yaml", ".yml", ".json":
		data, err := os.ReadFile(utils.ExpandHomeDirAndEnv(path))
		if err != nil {
			return err
		}

		return yaml.Unmarshal(data, out)
	default:
		return fmt.Errorf("unknown file extension: %s", path)
	}
}
