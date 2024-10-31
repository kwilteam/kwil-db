package app

import (
	"fmt"
	"strings"
)

var debugf = func(string, ...interface{}) {}

func enableCLIDebugging() {
	fmt.Println("CLI debugging enabled")
	debugf = func(msg string, args ...interface{}) {
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Printf("DEBUG: "+msg, args...)
	}
}

type lazyPrinter func() string

func (p lazyPrinter) String() string {
	return p()
}
