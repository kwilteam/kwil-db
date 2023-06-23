package utils

import (
	"fmt"
	"strings"

	ps "github.com/shirou/gopsutil/process"
)

func GetProcessID(pname string) int32 {
	processes, err := ps.Processes()
	if err != nil {
		return 0
	}

	for _, p := range processes {
		name, _ := p.Name()
		if strings.Contains(name, pname) {
			fmt.Printf("%s is running with pid %d\n", pname, p.Pid)
			return p.Pid
		}
	}
	fmt.Printf("%s is not running\n", pname)
	return 0
}
