//go:build js && wasm

package main

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/parser"
	"syscall/js"
)

func parse(input string) (json string, err error) {
	a, err := parser.Parse([]byte(input), parser.WithTraceOff())
	if err != nil {
		return "", err
	}

	json = string(a.Generate())
	return
}

// parseWrapper wraps the parse function to be exposed to the global scope
// returns a map {"json": "json output", "error": "error string"}
func parseWrapper() js.Func {
	parseFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		input := args[0].String()
		result := map[string]any{}

		jsonOut, err := parse(input)
		if err != nil {
			errStr := fmt.Sprintf("parsing failed: %s\n", err)
			result["error"] = errStr
		}
		result["json"] = jsonOut
		return result
	})
	return parseFunc
}

func main() {
	fmt.Println("Load KL parser...")
	// expose the parse function to the global scope
	js.Global().Set("parseKL", parseWrapper())
	<-make(chan bool)
}
