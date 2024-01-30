package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/optimizer/plan"
)

const (
	cliName = "ksql-planner"
)

// displayHelp informs the user about our hardcoded functions
func displayHelp(_ *bufio.Scanner) {
	fmt.Printf(
		"Welcome to %v! These are the available commands: \n",
		cliName,
	)
	fmt.Println(".help    - Show available commands")
}

func explain(reader *bufio.Scanner) {
	reader.Scan()
	text := cleanInput(reader.Text())
	plan.ExplainQuery(text)
}

func printPrompt() {
	fmt.Printf("%v> ", cliName)
}

// cleanInput preprocesses input to the repl
func cleanInput(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

type command func(*bufio.Scanner)

func main() {
	// Hardcoded repl commands
	commands := map[string]command{
		".help":    displayHelp,
		".explain": explain,
	}
	// Begin the repl loop
	reader := bufio.NewScanner(os.Stdin)
	printPrompt()
	for reader.Scan() {
		text := cleanInput(reader.Text())
		if command, exists := commands[text]; exists {
			command(reader)
		} else {
			fmt.Printf("Unknown command: %v\n", text)
		}
		printPrompt()
	}
	// Print an additional line if we encountered an EOF character
	fmt.Println()

}
