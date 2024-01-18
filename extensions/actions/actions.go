// package actions allows custom actions to be registered with the engine.
package actions

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/extensions"
)

var registeredExtensions = make(map[string]execution.ExtensionInitializer)

func RegisteredExtensions() map[string]execution.ExtensionInitializer {
	return registeredExtensions
}

// DEPRECATED: RegisterLegacyExtension registers an extension with the engine.
// It provides backwards compatibility with the old extension system.
// Use RegisterExtension instead.
func RegisterLegacyExtension(name string, ext extensions.LegacyEngineExtension) error {
	name = strings.ToLower(name)
	if _, ok := registeredExtensions[name]; ok {
		return fmt.Errorf("extension of same name already registered:%s ", name)
	}

	registeredExtensions[name] = extensions.AdaptLegacyExtension(ext)
	return nil
}
