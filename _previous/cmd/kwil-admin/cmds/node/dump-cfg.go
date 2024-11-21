package node

import (
	"context"
	"encoding/json"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

var (
	dumpCfgLong    = `Gets the current config from the node.`
	dumpCfgExample = `# Get the current config from the node.
kwil-admin node dump-config --rpcserver /tmp/kwild.socket`
)

func dumpCfgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dump-config",
		Short:   "Gets the current config from the node.",
		Long:    dumpCfgLong,
		Example: dumpCfgExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()
			client, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			bts, err := client.GetConfig(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			cfg := make(map[string]interface{})
			err = json.Unmarshal(bts, &cfg)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &cfgMsg{cfg: cfg})
		},
	}

	common.BindRPCFlags(cmd)

	return cmd
}

type cfgMsg struct {
	cfg map[string]interface{}
}

var _ display.MsgFormatter = (*cfgMsg)(nil)

func (c *cfgMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.cfg)
}

func (c *cfgMsg) MarshalText() ([]byte, error) {
	return toml.Marshal(c.cfg)
}
