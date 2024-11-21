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
	peersLong = `Print a list of the node's peers, with their public information.`

	peersExample = `# Print a list of the node's peers
kwil-admin node peers --rpcserver /tmp/kwild.socket`
)

func peersCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "peers",
		Short:   "Print a list of the node's peers, with their public information.",
		Long:    peersLong,
		Example: peersExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()
			client, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			peers, err := client.Peers(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
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
