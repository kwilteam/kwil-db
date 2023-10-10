package extensions

import "strings"

var registeredExtensions = make(map[string]*Extension)

func RegisterExtension(name string, ext *Extension) error {
	name = strings.ToLower(name)
	if _, ok := registeredExtensions[name]; ok {
		panic("extension of same name already registered: " + name)
	}

	registeredExtensions[name] = ext
	return nil
}

func GetRegisteredExtensions() map[string]*Extension {
	return registeredExtensions
}
