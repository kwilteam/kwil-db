package database

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/utils"
)

const (
	dbidFlag       = "dbid"
	nameFlag       = "name"
	ownerFlag      = "owner"
	actionNameFlag = "action"
)

// getSelectedOwner is used to get the owner flag.  Since the owner flag is usually optional,
// it will check to see if it was passed.  If it was not passed, it will attempt to
// get the user's public key from the configuration file.
func getSelectedOwner(cmd *cobra.Command, conf *config.KwilCliConfig) ([]byte, error) {
	var ident []byte
	if cmd.Flags().Changed(ownerFlag) {
		hexIdent, err := cmd.Flags().GetString(ownerFlag) // hex owner
		if err != nil {
			return nil, fmt.Errorf("failed to get public key from flag: %w", err)
		}

		// if it begins with 0x, remove it
		if len(hexIdent) > 2 && hexIdent[:2] == "0x" {
			hexIdent = hexIdent[2:]
		}

		ident, err = hex.DecodeString(hexIdent)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public key: %w", err)
		}

	} else {
		if conf.PrivateKey == nil {
			return nil, nil // nil is a valid owner, as it will return all
		}

		signer := auth.EthPersonalSigner{Key: *conf.PrivateKey}
		ident = signer.Identity()
	}

	return ident, nil
}

// getSelectedDbid returns the Dbid selected by the user.
// Since the user can pass either a name and owner, or a dbid, we need to
// check which one they passed and return the appropriate dbid.
// If only a name flag is passed, it will get the owner from the configuration file.
func getSelectedDbid(cmd *cobra.Command, conf *config.KwilCliConfig) (string, error) {
	if cmd.Flags().Changed(dbidFlag) {
		return cmd.Flags().GetString(dbidFlag)
	}

	if !cmd.Flags().Changed(nameFlag) {
		return "", fmt.Errorf("neither dbid nor name was provided")
	}

	name, err := cmd.Flags().GetString(nameFlag)
	if err != nil {
		return "", fmt.Errorf("failed to get name from flag: %w", err)
	}

	owner, err := getSelectedOwner(cmd, conf)
	if err != nil {
		return "", fmt.Errorf("failed to get owner flag: %w", err)
	}

	return utils.GenerateDBID(name, owner), nil
}

// bindFlagsTargetingProcedureOrAction binds the flags for any command that targets a procedure or action.
// This includes the `execute`, `call`, and `batch` commands.
func bindFlagsTargetingProcedureOrAction(cmd *cobra.Command) {
	bindFlagsTargetingDatabase(cmd)
	cmd.Flags().StringP(actionNameFlag, "a", "", "the target action name")
	err := cmd.Flags().MarkDeprecated(actionNameFlag, "pass the action name as the first argument")
	if err != nil {
		panic(err)
	}
}

// getSelectedActionOrProcedure returns the action or procedure name that the user selected.
// It is made to be backwards compatible with the old way of passing the action name as the --action flag.
// In v0.9, we changed this to have the action / procedure be passed as the first positional argument in
// all commands that require it.  This function will check if the --action flag was passed, and if it was,
// it will return that.  If it was not passed, it will return the first positional argument, and return the args
// with the first element removed.
func getSelectedActionOrProcedure(cmd *cobra.Command, args []string) (actionOrProc string, args2 []string, err error) {
	var actionOrProcedure string
	if cmd.Flags().Changed(actionNameFlag) {
		actionOrProcedure, err = cmd.Flags().GetString(actionNameFlag)
		if err != nil {
			return "", nil, err
		}
	} else {
		if len(args) < 1 {
			return "", nil, fmt.Errorf("missing action or procedure name")
		}

		actionOrProcedure = args[0]
		args = args[1:]
	}

	return strings.ToLower(actionOrProcedure), args, nil
}

// bindFlagsTargetingDatabase binds the flags for any command that targets a database.
// This includes the `query`, `execute`, `call`, and `batch` commands.
func bindFlagsTargetingDatabase(cmd *cobra.Command) {
	cmd.Flags().StringP(nameFlag, "n", "", "the target database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the target database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the target database id")
}
