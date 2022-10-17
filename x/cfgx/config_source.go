package cfgx

import (
	"fmt"
	"os"

	"kwil/x/utils"

	"gopkg.in/yaml.v2"
)

type configSourceImpl struct {
	name    string
	sources []SourceItem
}

func (c *configSourceImpl) Sources() []SourceItem {
	return c.sources
}

func createConfigSource(name string, sources ...SourceItem) Source {
	return &configSourceImpl{name: name, sources: sources}
}

func createConfigFileSource(path string) FileSource {
	return &configFileSourceImpl{path: path}
}

func createConfigFileSelectorSource(path string, selector string) FileSelectorSource {
	return &configFileSelectorSourceImpl{path: path, selector: selector}
}

func (c *configSourceImpl) add(source SourceItem) {
	c.sources = append(c.sources, source)
}

func (c *configSourceImpl) Name() string {
	return c.name
}

func (c *configSourceImpl) Items() []SourceItem {
	var local []SourceItem
	copy(local, c.sources)
	return local
}

type configFileSourceImpl struct {
	path string
}

func (c *configFileSourceImpl) Path() string {
	return c.path
}

type configFileSelectorSourceImpl struct {
	path     string
	selector string
}

func (c *configFileSelectorSourceImpl) Path() string {
	return c.path
}

func (c *configFileSelectorSourceImpl) Selector() string {
	return c.selector
}

func (c *configFileSourceImpl) As(out interface{}) error {
	return loadAs(c.path, out)
}

func (c *configFileSelectorSourceImpl) As(out interface{}) error {
	m := make(map[interface{}]interface{})
	err := loadAs(c.path, &m)
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

func loadAs(path string, out interface{}) error {
	data, err := os.ReadFile(utils.ExpandHomeDirAndEnv(path))
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, out)
}
