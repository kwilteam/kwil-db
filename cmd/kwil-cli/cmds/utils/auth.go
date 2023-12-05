package utils

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/gatewayclient"
	"github.com/spf13/cobra"
)

var authCmdDesc = `Authenticate with a Kwil Gateway using a private key.

The ` + `"` + `authenticate` + `"` + ` command will prompt you to sign a challenge from the Kwil Gateway. It
will store the returned auth cookie in ` + `"` + `~/.kwil-cli/auth` + `"` + ` for future use.

The Kwil CLI automatically handles authentication and re-authentication, however this tool
can be used to manually authenticate to a Kwil Gateway.
`

var authCmdExample = `# Authenticate to a Kwil Gateway

kwil-cli utils authenticate`

// kgwAuthnCmd is the command to authenticate to a KGW provider.
// This is not part of Kwil API.
func kgwAuthnCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "authenticate",
		Short:   "Authenticate with a Kwil Gateway using a private key.",
		Long:    authCmdDesc,
		Example: authCmdExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, common.UsingGateway,
				func(ctx context.Context, client common.Client, cfg *config.KwilCliConfig) error {
					if cfg.PrivateKey == nil {
						return display.PrintErr(cmd, fmt.Errorf("private key not provided"))
					}

					gatewayClient, ok := client.(*gatewayclient.GatewayClient)
					if !ok {
						return display.PrintErr(cmd, fmt.Errorf("client is not a gateway client. this is an internal bug"))
					}

					err := gatewayClient.Authenticate(ctx)
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("authentication failed: %w", err))
					}

					// we do not need to persist the cookie since DialClient will do that for us

					return display.PrintCmd(cmd, display.RespString("Success"))
				})
		},
	}

	return cmd
}
