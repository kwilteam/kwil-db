package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

func newQuitSignal() <-chan os.Signal {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	return s
}

var cliName string = "planREPL"

func printPrompt() {
	fmt.Print(cliName, "> ")
}

func formatInput(text string) string {
	output := strings.TrimSpace(text)
	output = strings.ToLower(output)
	return output
}

func explain() {
	fmt.Println("This is a stub for the explain command")
}

// builtin commands
var commands = map[string]func(){
	".help": func() {
		fmt.Printf("available commands: %s\n", cliName)
		fmt.Println(".help    - Show available commands")
		fmt.Println(".clear   - Clear the terminal screen")
		fmt.Println(".exit    - Closes the program")
		fmt.Println(".explain SQL- Explain the current query")
	},
	".exit": func() { os.Exit(0) },
	".clear": func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	},
	".explain": explain,
}

func handleSignal() {
	quitSignal := newQuitSignal()
	for {
		select {
		case <-quitSignal:
			fmt.Println("Received quit signal")
			os.Exit(0)
		}
	}
}

func main() {
	//go handleSignal()

	commands[".help"]()

	reader := bufio.NewScanner(os.Stdin)

	printPrompt()
	for reader.Scan() {
		text := formatInput(reader.Text())
		if command, exists := commands[text]; exists {
			command()
		} else {
			fmt.Println("Unknown command")
		}
		fmt.Println(strings.Repeat("-", 60))
		printPrompt()
	}

	fmt.Println()
}
