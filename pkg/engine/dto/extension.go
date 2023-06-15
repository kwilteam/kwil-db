package dto

import "strings"

type ExtensionInitialization struct {
	Name     string
	Metadata map[string]string
}

func (e *ExtensionInitialization) Clean() error { // returns an error to make it consistent with other dto.Clean() methods
	e.Name = strings.ToLower(e.Name)
	return nil
}
