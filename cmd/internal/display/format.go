// Package display provides interfaces and functions to format the command
// line output and print.
package display

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
)

// emptyResult is used as a placeholder for when the result is empty.
type emptyResult struct{}

func (e *emptyResult) MarshalText() ([]byte, error) {
	return []byte(""), nil
}

func (e *emptyResult) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}

type msgFormatter interface {
	json.Marshaler
	encoding.TextMarshaler
}

type msgPrinter interface {
	printJson(stdout io.Writer, stderr io.Writer) error
	printText(stdout io.Writer, stderr io.Writer) error
}

type wrappedMsg struct {
	Result msgFormatter `json:"result"`
	Error  string       `json:"error"`
}

// printJson prints the wrappedMsg in json format. It prints to stdout.
func (w *wrappedMsg) printJson(stdout io.Writer, _ io.Writer) error {
	msg, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, string(msg))
	return nil
}

// printText prints the wrappedMsg in text format. It prints to stdout if
// `w.Error` is empty, otherwise it prints to stderr.
func (w *wrappedMsg) printText(stdout io.Writer, stderr io.Writer) error {
	if w.Error != "" {
		fmt.Fprintln(stderr, w.Error)
		return nil
	}

	msg, err := w.Result.MarshalText()
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, string(msg))
	return nil
}

// WrapMsg wraps response and error in a wrappedMsg struct.
func WrapMsg(msg msgFormatter, err error) *wrappedMsg {
	if err != nil {
		return &wrappedMsg{
			Result: &emptyResult{},
			Error:  err.Error(),
		}
	}

	return &wrappedMsg{
		Result: msg,
		Error:  "",
	}
}

func PrettyPrint(pt msgPrinter, format string, stdout io.Writer, stderr io.Writer) error {
	switch format {
	case config.OutputFormatJSON.String():
		return pt.printJson(stdout, stderr)
	default:
		return pt.printText(stdout, stderr)

	}
}

// Print is a helper function to wrap and print message in given format.
func Print(msg msgFormatter, err error, format string) error {
	wrappedMsg := WrapMsg(msg, err)
	return PrettyPrint(wrappedMsg, format, os.Stdout, os.Stderr)
}
