//go:build js && wasm

//go:generate env GOOS=js CGO_ENABLED=0 GOARCH=wasm go build -o kuneiform.wasm -ldflags "-s -w" -trimpath -tags netgo wasm.go
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/kwilteam/kwil-db/parse"
)

func parseAndMarshal(input string) (jsonStr string, err error) {
	schema, err := parse.ParseAndValidate([]byte(input))
	if err != nil {
		return "", err
	}

	// remove parsed action and procedure asts
	schema.ParsedActions = nil
	schema.ParsedProcedures = nil

	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}

	jsonStr = string(jsonBytes)
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

		jsonOut, err := parseAndMarshal(input)
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
	fmt.Println("Loading Kuneiform parser...")
	// expose the parse function to the global scope
	js.Global().Set("parseKuneiform", parseWrapper())
	<-make(chan bool)
}
