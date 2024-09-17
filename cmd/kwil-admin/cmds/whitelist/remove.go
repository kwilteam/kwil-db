package whitelist

import (
	"context"
	"encoding/json"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

func removeCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "remove <peerID>",
		Short:   "Remove a peer from the node's whitelist peers and disconnect the peer",
		Example: "kwil-admin whitelist remove <peerID>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			err = client.RemovePeer(ctx, args[0])
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &removeMsg{peerID: args[0]})
		},
	}
	common.BindRPCFlags(cmd)

	return cmd
}

type removeMsg struct {
	peerID string
}

var _ display.MsgFormatter = (*addMsg)(nil)

func (a *removeMsg) MarshalText() ([]byte, error) {
	return []byte("Removed peer " + a.peerID), nil
}

func (a *removeMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.peerID)
}
