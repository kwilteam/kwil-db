package database

import (
	"fmt"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/crypto"
	"kwil/pkg/engine/models"

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
// get the user's address from the configuration file.
func getSelectedOwner(cmd *cobra.Command, conf *config.KwilCliConfig) (string, error) {
	var address string
	if cmd.Flags().Changed(ownerFlag) {
		var err error
		address, err = cmd.Flags().GetString(ownerFlag)
		if err != nil {
			return address, fmt.Errorf("failed to get address from flag: %w", err)
		}

		if address == "" {
			return address, fmt.Errorf("no address provided")
		}

		if !crypto.IsValidAddress(address) {
			return address, fmt.Errorf("invalid address provided: %s", address)
		}
	} else {
		address = crypto.AddressFromPrivateKey(conf.PrivateKey)
	}

	return address, nil
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

	return models.GenerateSchemaId(owner, name), nil
}
