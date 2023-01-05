package executables

import "kwil/x/execution/dto"

func (d *executableInterface) ListExecutables() []*dto.Executable {
	var list []*dto.Executable
	for _, v := range d.Executables {
		list = append(list, v)
	}
	return list
}
