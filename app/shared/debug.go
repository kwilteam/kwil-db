package shared

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var Debugf = func(string, ...interface{}) {}

func EnableCLIDebugging() {
	fmt.Println("CLI debugging enabled")
	Debugf = func(msg string, args ...interface{}) {
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Printf("DEBUG: "+msg, args...)
	}
}

type LazyPrinter func() string

func (p LazyPrinter) String() string {
	return p()
}

func BindDebugFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("debug", false, "enable debugging, will print debug logs")
	cmd.Flag("debug").Hidden = true
}

func MaybeEnableCLIDebug(cmd *cobra.Command, args []string) error {
	debugFlag := cmd.Flag("debug")
	if !debugFlag.Changed {
		return nil
	}
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return err
	}
	if debug {
		EnableCLIDebugging()
		k.Set("log_level", "debug")
	}
	return nil
}
