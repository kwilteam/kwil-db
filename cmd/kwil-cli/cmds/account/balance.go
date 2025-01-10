package account

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/client"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/spf13/cobra"
)

func balanceCmd() *cobra.Command {
	var pending bool
	cmd := &cobra.Command{
		Use:   "balance accountID keyType",
		Short: "Gets an account's balance and nonce",
		Long:  `Gets an account's balance and nonce.`,
		Args:  cobra.MaximumNArgs(2), // no args means own account
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var acctID *types.AccountID
			var typeStr string
			var clientFlags uint8

			if len(args) > 0 {
				clientFlags = client.WithoutPrivateKey

				id, err := hex.DecodeString(args[0])
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("failed to decode account ID: %w", err))
				}

				if len(args) == 1 {
					typeStr = crypto.KeyTypeSecp256k1.String()
				} else {
					typeStr = args[1]
				}

				keyType, err := crypto.ParseKeyType(typeStr)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("failed to parse key type %s: %w", typeStr, err))
				}

				acctID = &types.AccountID{
					Identifier: id,
					KeyType:    keyType,
				}
			} // else use our account from the signer

			return client.DialClient(cmd.Context(), cmd, clientFlags, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				if len(args) == 0 {
					if cl.Signer() == nil {
						return display.PrintErr(cmd, errors.New("no account ID provided and no signer set"))
					}

					acctID, err = types.GetSignerAccount(cl.Signer())
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("failed to get signer account: %w", err))
					}

				}
				status := types.AccountStatusLatest
				if pending {
					status = types.AccountStatusPending
				}
				acct, err := cl.GetAccount(ctx, acctID, status)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("get account failed: %w", err))
				}
				// NOTE: empty acct.Identifier means it doesn't even have a record
				// on the network. Perhaps we convey that to the caller? Their
				// balance is zero regardless, assuming it's the correct acct ID.

				resp := &respAccount{
					Balance: acct.Balance.String(),
					Nonce:   acct.Nonce,
				}

				if acct.ID != nil { // only add identifier for the existing accounts
					resp.Identifier = acct.ID.Identifier
					resp.KeyType = acct.ID.KeyType.String()
				}

				return display.PrintCmd(cmd, resp)
			})

		},
	}

	cmd.Flags().BoolVar(&pending, "pending", false, "reflect pending updates from mempool (default is confirmed only)")

	return cmd
}
