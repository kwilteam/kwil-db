package actions

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types/extensions"
)

var registeredExtensions = make(map[string]extensions.EngineExtension)

func RegisterExtension(name string, ext extensions.EngineExtension) error {
	name = strings.ToLower(name)
	if _, ok := registeredExtensions[name]; ok {
		return fmt.Errorf("extension of same name already registered:%s ", name)
	}

	registeredExtensions[name] = ext
	return nil
}

func RegisteredExtensions() map[string]extensions.EngineExtension {
	return registeredExtensions
}
