package shared

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

type CmdCtxKey string

var (
	CmdCtxKeyCmdErr CmdCtxKey = "cmdErr"
)

// SetCmdCtxErr records the error in the command's context. Use CmdCtxErr to
// extract the error.
func SetCmdCtxErr(cmd *cobra.Command, err error) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	if ctxErr, _ := ctx.Value(CmdCtxKeyCmdErr).(error); ctxErr != nil {
		err = errors.Join(err, ctxErr)
	}
	ctx = context.WithValue(ctx, CmdCtxKeyCmdErr, err)
	cmd.SetContext(ctx)
}

// CmdCtxErr returns the error stored in the command's context with SetCmdCtxErr.
func CmdCtxErr(cmd *cobra.Command) error {
	ctx := cmd.Context()
	if ctx == nil {
		return nil
	}
	ctxErr, _ := ctx.Value(CmdCtxKeyCmdErr).(error)
	return ctxErr
}

func removeBackticks(s string) string {
	return strings.ReplaceAll(s, "`", "'")
}

// SetSanitizedHelpFunc remove backticks from the help text when printing to the
// console. The backticks are there for generated Docusaurus documentation, but
// they are odd for a user reading the help text in the console.
func SetSanitizedHelpFunc(cmd *cobra.Command) {
	// the default HelpFunc that we will delegate back to after revising the help text
	originalHelpFunc := cmd.HelpFunc()

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Short = wrapTextToTerminalWidth(removeBackticks(cmd.Short))
		cmd.Long = wrapTextToTerminalWidth(removeBackticks(cmd.Long))
		wrapFlags(cmd.Flags())
		wrapFlags(cmd.PersistentFlags())

		// Delegate to the original HelpFunc to avoid recursion
		if originalHelpFunc != nil {
			originalHelpFunc(cmd, args)
		} else {
			// If there's no original HelpFunc, use the usage function as a fallback
			if usageFunc := cmd.UsageFunc(); usageFunc != nil {
				usageFunc(cmd)
			}
		}
	})
}

func ApplySanitizedHelpFuncRecursively(cmd *cobra.Command) {
	SetSanitizedHelpFunc(cmd)

	for _, subCmd := range cmd.Commands() {
		ApplySanitizedHelpFuncRecursively(subCmd)
	}
}
