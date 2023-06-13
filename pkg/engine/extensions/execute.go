package extensions

import (
	"fmt"
	"os/exec"
)

func (e *Extension) ExportedFunction(name string) (*ExtensionFunction, error) {
	fn, ok := e.Functions[name]
	if !ok {
		return nil, fmt.Errorf("function %s not found", name)
	}

	return fn, nil
}

func (e *ExtensionFunction) Execute(args ...any) ([]any, error) {

	cm := exec.Command()
}
