package validators

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/internal/validators"
	"github.com/spf13/cobra"
)

var (
	listLong = `List the current validator set of the network.`

	listExample = `# List the current validator set of the network
kwild validator list`
)

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List the current validator set of the network.",
		Long:    listLong,
		Example: listExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return err
			}

			data, err := clt.ListValidators(ctx)
			if err != nil {
				return err
			}

			return display.PrintCmd(cmd, &respValSets{Data: data})
		},
	}

	return cmd
}

// respValSets represent current validator set in cli
type respValSets struct {
	Data []*validators.Validator
}

type valInfo struct {
	PubKey string `json:"pubkey"`
	Power  int64  `json:"power"`
}

func (r *respValSets) MarshalJSON() ([]byte, error) {
	valInfos := make([]valInfo, len(r.Data))
	for i, v := range r.Data {
		valInfos[i] = valInfo{
			PubKey: fmt.Sprintf("%x", v.PubKey),
			Power:  v.Power,
		}
	}

	return json.Marshal(valInfos)
}

func (r *respValSets) MarshalText() ([]byte, error) {
	var msg bytes.Buffer
	msg.WriteString("Current validator set:\n")
	for i, v := range r.Data {
		msg.WriteString(fmt.Sprintf("% 3d. %s\n", i, v))
	}

	return msg.Bytes(), nil
}
