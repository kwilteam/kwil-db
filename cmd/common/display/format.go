// Package display provides interfaces and functions to format the command
// line output and print.
package display

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// BindOutputFormatFlag binds the output format flag to the command.
// This should be added on the root command.
func BindOutputFormatFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String("output", defaultOutputFormat.string(), "the format for command output - either 'text' or 'json'")
}

// BindSilenceFlag binds the silence flag to the passed command.
// If bound, the command will silence logs.
// If true, display commands will not print to stdout or stderr.
// The flag will be bound to all subcommands of the given command.
func BindSilenceFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP("silence", "S", false, "Silence logs")
}

// ShouldSilence returns the value of the silence flag
func ShouldSilence(cmd *cobra.Command) bool {
	s, _ := cmd.Flags().GetBool("silence")
	return s
}

// OutputFormat is the format for command output
// It implements the pflag.Value interface
type OutputFormat string

// String implements the Stringer interface
// NOTE: cannot use the pointer receiver here
func (o OutputFormat) string() string {
	return string(o)
}

// Valid returns true if the output format is valid.
func (o OutputFormat) valid() bool {
	switch o {
	case outputFormatText, outputFormatJSON:
		return true
	default:
		return false
	}
}

const (
	outputFormatText OutputFormat = "text"
	outputFormatJSON OutputFormat = "json"

	defaultOutputFormat = outputFormatText
)

// emptyResult is used as a placeholder for when the result is empty.
// It implements the MsgFormatter interface
type emptyResult struct{}

func (e *emptyResult) MarshalText() ([]byte, error) {
	return []byte(""), nil
}

func (e *emptyResult) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}

// MsgFormatter is an interface that wraps the MarshalText and MarshalJSON
// It defines the requirements for something to be printed.
type MsgFormatter interface {
	json.Marshaler
	encoding.TextMarshaler
}

type wrappedMsg struct {
	Result MsgFormatter `json:"result"`
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

// wrapMsg wraps response and error in a wrappedMsg struct.
func wrapMsg(msg MsgFormatter, err error) *wrappedMsg {
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

func prettyPrint(msg *wrappedMsg, format OutputFormat, stdout io.Writer, stderr io.Writer) error {
	switch format {
	case outputFormatJSON:
		return msg.printJson(stdout, stderr)
	case outputFormatText:
		return msg.printText(stdout, stderr)
	default:
		return fmt.Errorf("invalid output format: %s", format)
	}
}

// Print is a helper function to wrap and print message in given format.
func Print(msg MsgFormatter, err error, format OutputFormat) error {
	wrappedMsg := wrapMsg(msg, err)
	return prettyPrint(wrappedMsg, format, os.Stdout, os.Stderr)
}

// PrintCmd prints output based on the commands output format flag.
// If not format flag is provided, it will default to text in stdout.
func PrintCmd(cmd *cobra.Command, msg MsgFormatter) error {
	if ShouldSilence(cmd) {
		return nil
	}

	wrappedMsg := &wrappedMsg{
		Result: msg,
		Error:  "",
	}

	format, err := cmd.Flags().GetString("output")
	if err != nil || format == "" {
		format = defaultOutputFormat.string()
	}
	if !OutputFormat(format).valid() {
		// set the output format in cmd to the default output format to avoid error being thrown twice
		err := cmd.Flags().Set("output", defaultOutputFormat.string())
		if err != nil {
			return err
		}

		return fmt.Errorf("invalid output format: %s", format)
	}

	return prettyPrint(wrappedMsg, OutputFormat(format), os.Stdout, os.Stderr)
}

// PrintErr prints the error according to the commands output format flag.
func PrintErr(cmd *cobra.Command, err error) error {
	if ShouldSilence(cmd) {
		return nil
	}

	outputFormat, err2 := getOutputFormat(cmd)
	if err2 != nil {
		return err2
	}

	return prettyPrint(&wrappedMsg{
		Result: &emptyResult{},
		Error:  err.Error(),
	}, outputFormat, os.Stdout, os.Stderr)
}

func getOutputFormat(cmd *cobra.Command) (OutputFormat, error) {
	format, err := cmd.Flags().GetString("output")
	if err != nil || format == "" {
		format = defaultOutputFormat.string()
	}
	if !OutputFormat(format).valid() {
		return "", fmt.Errorf("invalid output format: %s", format)
	}

	return OutputFormat(format), nil
}
