package whitelist

import (
	"context"
	"encoding/json"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "add <peerID>",
		Short:   "Add a peer to the node's whitelist peers to accept connections from.",
		Example: "kwil-admin whitelist add <peerID>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			err = client.AddPeer(ctx, args[0])
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &addMsg{peerID: args[0]})
		},
	}
	common.BindRPCFlags(cmd)

	return cmd
}

type addMsg struct {
	peerID string
}

var _ display.MsgFormatter = (*addMsg)(nil)

func (a *addMsg) MarshalText() ([]byte, error) {
	return []byte("Whitelisted peer " + a.peerID), nil
}

func (a *addMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.peerID)
}
