package utils

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var authCmdDesc = `KGW provider provides ways to protect data privacy, by using cookie authentication.
This command will prompt for a signature and return a authenticated cookie for
future API calls.
KGW authentication is not part of Kwild API.
`

// kgwAuthnCmd is the command to authenticate to a KGW provider.
// This is not part of Kwil API.
func kgwAuthnCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "kgw-authn",
		Short: "kgw-authn is used to do authentication with a KGW provider", // or sass provider?
		Long:  authCmdDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := common.DialClient(cmd.Context(), 0,
				func(ctx context.Context, client *client.Client,
					cfg *config.KwilCliConfig) error {
					if cfg.PrivateKey == nil {
						return fmt.Errorf("private key not provided")
					}

					signer := auth.EthPersonalSigner{Key: *cfg.PrivateKey}
					if cfg.GrpcURL == "" {
						return fmt.Errorf("provider url not provided")
					}

					userAddress, err := signer.Address()
					if err != nil {
						return fmt.Errorf("get address: %w", err)
					}

					cookie, err := client.KGWAuthenticate(ctx, promptMessage)
					if err != nil {
						return fmt.Errorf("KGW authenticate: %w", err)
					}

					err = common.SaveAuthInfo(common.KGWAuthTokenFilePath(),
						userAddress, cookie)
					if err != nil {
						return fmt.Errorf("save auth token: %w", err)
					}

					return nil
				})

			return display.Print(respStr("Success"), err, config.GetOutputFormat())
		},
	}

	return cmd
}

// promptMessage prompts the user to sign a message. Return an error if user
// declines to sign.
func promptMessage(msg string) error {
	// display the message to user
	fmt.Println(msg)

	prompt := promptui.Prompt{
		Label:     "Do you want to sign this message?",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("you declined to sign")
	}

	return nil
}
