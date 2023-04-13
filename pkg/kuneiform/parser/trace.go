package parser

import (
	"fmt"
	"strings"
)

const traceBegan = ">> "
const traceEnded = "<< "
const indentPlaceholder = "\t"

var indent int = 0 // indentation

func indentPrint(msg string) {
	fmt.Printf("%s%s\n", strings.Repeat(indentPlaceholder, indent-1), msg)
}

func trace(msg string) string {
	indent++
	indentPrint(traceBegan + msg)
	return msg
}

func un(msg string) {
	indentPrint(traceEnded + msg)
	indent--
}
