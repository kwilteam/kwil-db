package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const (
	nameFlag       = "namespace"
	actionNameFlag = "action"
)

// getSelectedNamespace returns the Dbid selected by the user.
// Since the user can pass either a name and owner, or a dbid, we need to
// check which one they passed and return the appropriate dbid.
// If only a name flag is passed, it will get the owner from the configuration file.
func getSelectedNamespace(cmd *cobra.Command) (string, error) {
	if !cmd.Flags().Changed(nameFlag) {
		return "", errors.New("no namespace selected")
	}

	name, err := cmd.Flags().GetString(nameFlag)
	if err != nil {
		return "", fmt.Errorf("failed to get name from flag: %w", err)
	}

	return name, nil
}

// bindFlagsTargetingAction binds the flags for any command that targets a procedure or action.
// This includes the `execute`, `call`, and `batch` commands.
func bindFlagsTargetingAction(cmd *cobra.Command) {
	cmd.Flags().StringP(nameFlag, "n", "", "the target database namespace")
	cmd.Flags().StringP(actionNameFlag, "a", "", "the target action name")
}

// getSelectedAction returns the action or procedure name that the user selected.
// It is made to be backwards compatible with the old way of passing the action name as the --action flag.
// In v0.9, we changed this to have the action / procedure be passed as the first positional argument in
// all commands that require it.  This function will check if the --action flag was passed, and if it was,
// it will return that.  If it was not passed, it will return the first positional argument, and return the args
// with the first element removed.
func getSelectedAction(cmd *cobra.Command, args []string) (action string, args2 []string, err error) {
	var actionOrProcedure string
	if actionFlagSet(cmd) {
		actionOrProcedure, err = cmd.Flags().GetString(actionNameFlag)
		if err != nil {
			return "", nil, err
		}
	} else {
		if len(args) < 1 {
			return "", nil, errors.New("missing action name or SQL statement")
		}

		actionOrProcedure = args[0]
		args = args[1:]
	}

	return strings.ToLower(actionOrProcedure), args, nil
}

func actionFlagSet(cmd *cobra.Command) bool {
	return cmd.Flags().Changed(actionNameFlag)
}
