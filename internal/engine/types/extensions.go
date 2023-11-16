package types

import "fmt"

type Extension struct {
	Name           string             `json:"name"`
	Initialization []*ExtensionConfig `json:"initialization"`
	Alias          string             `json:"alias"`
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (e *Extension) Clean() error {
	keys := make(map[string]struct{})
	for _, config := range e.Initialization {
		_, ok := keys[config.Key]
		if ok {
			return fmt.Errorf("duplicate key %s in extension %s", config.Key, e.Name)
		}

		keys[config.Key] = struct{}{}
	}

	return runCleans(
		cleanIdent(&e.Name),
		cleanIdent(&e.Alias),
	)
}

// ConfigMap returns a map of the config values for the extension
func (e *Extension) ConfigMap() map[string]string {
	config := make(map[string]string)
	for _, c := range e.Initialization {
		config[c.Key] = c.Value
	}

	return config
}

// ExtensionConfig is a key value pair that represents a configuration value for an extension
type ExtensionConfig struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}
