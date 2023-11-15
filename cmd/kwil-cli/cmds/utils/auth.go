package utils

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/gatewayclient"
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
		Short: "kgw-authn is used to do authentication with a KGW provider",
		Long:  authCmdDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, common.UsingGateway,
				func(ctx context.Context, client common.Client, cfg *config.KwilCliConfig) error {
					if cfg.PrivateKey == nil {
						return fmt.Errorf("private key not provided")
					}

					gatewayClient, ok := client.(*gatewayclient.GatewayClient)
					if !ok {
						return fmt.Errorf("client is not a gateway client. this is an internal bug")
					}

					err := gatewayClient.Authenticate(ctx)
					if err != nil {
						return fmt.Errorf("authentication failed: %w", err)
					}

					// retrieve the cookie and persist it
					cookie, found := gatewayClient.GetAuthCookie()
					if !found {
						return fmt.Errorf("authentication failed: cookie could not be found")
					}

					err = common.SaveCookie(common.KGWAuthTokenFilePath(), gatewayClient.Signer.Identity(), cookie)
					if err != nil {
						return fmt.Errorf("save cookie: %w", err)
					}

					return display.PrintCmd(cmd, display.RespString("Success"))
				})
		},
	}

	return cmd
}
