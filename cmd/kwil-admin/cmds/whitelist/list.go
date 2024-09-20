package whitelist

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "list",
		Short:   "List the peers in the node's whitelist.",
		Example: "kwil-admin whitelist list",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			peers, err := client.ListPeers(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &listPeersMsg{peers: peers})
		},
	}
	common.BindRPCFlags(cmd)

	return cmd
}

type listPeersMsg struct {
	peers []string
}

var _ display.MsgFormatter = (*listPeersMsg)(nil)

func (l *listPeersMsg) MarshalText() ([]byte, error) {
	return []byte("Whitelisted Peers:  \n" + strings.Join(l.peers, "\n")), nil
}

func (l *listPeersMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.peers)
}
