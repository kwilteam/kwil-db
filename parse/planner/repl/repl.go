package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/kwilteam/kwil-db/parse/planner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run repl.go <kuneiform_filename>")
		return
	}

	filename := os.Args[1]

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	schema, err := parse.Parse(fileContent)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Kuneiform Logical Planner REPL")
	fmt.Println("Type 'exit' to exit")
	fmt.Println("-----------------------------")

	for {
		fmt.Print(">> ")
		text, _ := reader.ReadString(';')

		printLogical(schema, text)
	}
}

func printLogical(schema *types.Schema, sql string) {
	parsed, err := parse.ParseSQL(sql, schema, true)
	if err != nil {
		fmt.Printf("Error parsing SQL: %v\n", err)
		return
	}
	if parsed.ParseErrs.Err() != nil {
		fmt.Printf("Error parsing SQL: %v\n", parsed.ParseErrs.Err())
		return
	}

	plan, err := planner.Plan(parsed.AST, schema, nil, nil)
	if err != nil {
		fmt.Printf("Error planning SQL: %v\n", err)
		return
	}

	fmt.Println(planner.Format(plan))
}
