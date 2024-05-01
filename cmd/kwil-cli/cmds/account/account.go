package account

import (
	"encoding/hex"
	"errors"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/crypto/auth"

	"github.com/spf13/cobra"
)

var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Show the account ID.",
	Long:  "Returns the Kwil account identifier (typically an Ethereum address), if a private key is configured.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		conf, err := config.LoadCliConfig()
		if err != nil {
			return display.PrintErr(cmd, err)
		}

		if conf.PrivateKey == nil {
			return display.PrintErr(cmd, errors.New("no private key configured"))
		}

		signer := &auth.EthPersonalSigner{Key: *conf.PrivateKey}
		return display.PrintCmd(cmd, display.RespString(hex.EncodeToString(signer.Identity())))
	},
}

var (
	nonceOverride int64
	syncBcast     bool
)

func NewCmdAccount() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "account",
		Short: "Account related commands.",
		Long:  "Commands related to Kwil account, such as balance checks and transfers.",
	}

	trCmd := transferCmd() // gets the nonce override flag

	cmd.AddCommand(
		idCmd,
		balanceCmd(),
		trCmd,
	)

	trCmd.Flags().Int64VarP(&nonceOverride, "nonce", "N", -1, "nonce override (-1 means request from server)")
	trCmd.Flags().BoolVar(&syncBcast, "sync", false, "synchronous broadcast (wait for it to be included in a block)")

	return cmd
}
