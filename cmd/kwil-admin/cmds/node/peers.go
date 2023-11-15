package node

import (
	"context"
	"encoding/json"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/spf13/cobra"
)

var (
	peersLong = `Print a list of the node's peers, with their public information.`

	peersExample = `# Print a list of the node's peers
kwil-admin node peers --rpcserver localhost:50151 --authrpc-cert "~/.kwild/rpc.cert"`
)

func peersCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "peers",
		Short:   "Print a list of the node's peers, with their public information.",
		Long:    peersLong,
		Example: peersExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return err
			}

			peers, err := client.Peers(ctx)
			if err != nil {
				return err
			}

			return display.PrintCmd(cmd, &peersMsg{peers: peers})
		},
	}

	common.BindRPCFlags(cmd)

	return cmd
}

// peersMsg is a wrapper around the []*types.PeerInfo type that
// implements the MsgFormatter interface.
type peersMsg struct {
	peers []*types.PeerInfo
}

var _ display.MsgFormatter = (*peersMsg)(nil)

func (p *peersMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.peers)
}

func (p *peersMsg) MarshalText() ([]byte, error) {
	return json.MarshalIndent(p.peers, "", "  ")
}
