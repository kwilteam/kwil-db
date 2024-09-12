package database

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/spf13/cobra"
)

const (
	dbidFlag       = "dbid"
	nameFlag       = "name"
	ownerFlag      = "owner"
	actionNameFlag = "action"
	targetFlag     = "target"
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
	cmd.Flags().MarkDeprecated(actionNameFlag, "please use --target instead")
	cmd.Flags().StringP(targetFlag, "t", "", "the target action or procedure name")
}

func getSelectedProcedureAndDBID(cmd *cobra.Command, conf *config.KwilCliConfig) (dbid string, procOrAction string, err error) {
	dbid, err = getSelectedDbid(cmd, conf)
	if err != nil {
		return "", "", fmt.Errorf("failed to get dbid: %w", err)
	}

	var name string
	if cmd.Flags().Changed(targetFlag) {
		name, err = cmd.Flags().GetString(targetFlag)
		if err != nil {
			return "", "", fmt.Errorf("failed to get procedure name: %w", err)
		}
	} else if cmd.Flags().Changed(actionNameFlag) {
		name, err = cmd.Flags().GetString(actionNameFlag)
		if err != nil {
			return "", "", fmt.Errorf("failed to get action name: %w", err)
		}
	} else {
		return "", "", fmt.Errorf("neither procedure nor action was provided")
	}

	return dbid, strings.ToLower(name), nil
}

// bindFlagsTargetingDatabase binds the flags for any command that targets a database.
// This includes the `query`, `execute`, `call`, and `batch` commands.
func bindFlagsTargetingDatabase(cmd *cobra.Command) {
	cmd.Flags().StringP(nameFlag, "n", "", "the target database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the target database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the target database id")
}
