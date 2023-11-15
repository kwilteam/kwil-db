package common

import "github.com/spf13/cobra"

// this file can be used to define flags that should be globally accessible / shared between commands

// BindAssumeYesFlag binds the assume yes flag to the passed command
// If bound, the command will assume yes for all prompts
func BindAssumeYesFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("assume-yes", "Y", false, "Assume yes for all prompts")
}

// GetAssumeYesFlag returns the value of the assume yes flag
func GetAssumeYesFlag(cmd *cobra.Command) (bool, error) {
	return cmd.Flags().GetBool("assume-yes")
}
