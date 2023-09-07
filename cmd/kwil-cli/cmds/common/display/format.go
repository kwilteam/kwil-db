package display

import (
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
)

func PrettyPrint(pt msgPrinter, format string) error {
	switch format {
	case config.OutputFormatJSON.String():
		return pt.printJson()
	default:
		return pt.printText()
	}
}

type msgFormatter interface {
	MarshalJSON() ([]byte, error)
	MarshalText() (string, error)
}

type msgPrinter interface {
	printJson() error
	printText() error
}

type wrappedMsg struct {
	Result msgFormatter `json:"result"`
	Error  string       `json:"error"`
}

func (w *wrappedMsg) printJson() error {
	msg, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(msg))
	return nil
}

func (w *wrappedMsg) printText() error {
	msg, err := w.Result.MarshalText()
	if err != nil {
		return err
	}

	fmt.Println(msg)
	return nil
}

// WrapMsg wraps response and error in a wrappedMsg struct.
func WrapMsg(msg msgFormatter, err error) *wrappedMsg {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	return &wrappedMsg{
		Result: msg,
		Error:  errMsg,
	}
}

// Print is a helper function to short circuit on error.
// It then prints the message using printer in the desired format.
func Print(printer msgPrinter, err error, format string) error {
	// short circuit.
	// NOTE: we can also put error in FormatText, but i think it would break the
	// expected behavior of FormatText
	if format == config.OutputFormatText.String() && err != nil {
		return err
	}

	return PrettyPrint(printer, format)
}
