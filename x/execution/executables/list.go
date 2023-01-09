package executables

import (
	execTypes "kwil/x/types/execution"
)

func (d *executableInterface) ListExecutables() []*execTypes.Executable {
	var list []*execTypes.Executable
	for _, v := range d.Executables {
		list = append(list, v)
	}
	return list
}
