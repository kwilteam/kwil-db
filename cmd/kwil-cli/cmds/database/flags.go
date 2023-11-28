package database

import (
	"encoding/hex"
	"fmt"

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
			return nil, fmt.Errorf("no public key provided")
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
