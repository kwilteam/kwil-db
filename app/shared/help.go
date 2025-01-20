package shared

import (
	"strings"

	"github.com/spf13/cobra"
)

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
		cmd.Short = removeBackticks(cmd.Short)
		cmd.Long = removeBackticks(cmd.Long)

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
