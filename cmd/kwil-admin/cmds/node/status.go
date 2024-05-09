package node

import (
	"context"
	"encoding/json"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/spf13/cobra"
)

var (
	statusLong = `Print the node's status information.`

	statusExample = `# Print a remote node's status (TLS with self-signed certificate)
kwil-admin node status --rpcserver /tmp/kwild.socket`
)

func statusCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "status",
		Short:   "Print the node's status information.",
		Long:    statusLong,
		Example: statusExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()
			client, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			status, err := client.Status(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &statusMsg{status: status})
		},
	}

	common.BindRPCFlags(cmd)

	return cmd
}

// statusMsg is a wrapper around the Status type that
// implements the MsgFormatter interface.
type statusMsg struct {
	status *types.Status
}

var _ display.MsgFormatter = (*statusMsg)(nil)

func (s *statusMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.status)
}

func (s *statusMsg) MarshalText() ([]byte, error) {
	return json.MarshalIndent(s.status, "", "  ")
}
