package node

import (
	"context"
	"encoding/json"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

var (
	dumpCfgLong    = `Gets the current config from the node.`
	dumpCfgExample = `# Get the current config from the node.
kwil node dump-config --rpcserver unix:///tmp/kwil_admin.sock`
)

func dumpCfgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dump-config",
		Short:   "Gets the current config from the node.",
		Long:    dumpCfgLong,
		Example: dumpCfgExample,
		RunE: func(cmd *cobra.Command, args []string) error {
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
	maps := make(map[string]interface{})
	err := mapstructure.Decode(c.cfg, &maps)
	if err != nil {
		return nil, err
	}

	return json.Marshal(maps)
}

func (c *cfgMsg) MarshalText() ([]byte, error) {
	maps := make(map[string]interface{})
	err := mapstructure.Decode(c.cfg, &maps)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(maps, "", "  ")
}
