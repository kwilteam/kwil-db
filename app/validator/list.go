package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/core/types"
)

var (
	listLong = `List the current validator set of the network.`

	listExample = `# List the current validator set of the network
kwil-admin validators list`
)

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",                      // validators list
		Aliases: []string{"list-validators"}, // validators list-validators
		Short:   "List the current validator set of the network.",
		Long:    listLong,
		Example: listExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := rpc.AdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			data, err := clt.ListValidators(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &respValSets{Data: data})
		},
	}

	return cmd
}

// respValSets represent current validator set in cli
type respValSets struct {
	Data []*types.Validator
}

type valInfo struct {
	PubKey     string `json:"pubkey"`
	PubKeyType string `json:"pubkey_type"`
	Power      int64  `json:"power"`
}

func (r *respValSets) MarshalJSON() ([]byte, error) {
	valInfos := make([]*valInfo, len(r.Data))
	for i, v := range r.Data {
		valInfos[i] = &valInfo{
			PubKey:     v.Identifier.String(),
			Power:      v.Power,
			PubKeyType: v.KeyType.String(),
		}
	}

	return json.Marshal(valInfos)
}

func (r *respValSets) MarshalText() ([]byte, error) {
	var msg bytes.Buffer
	msg.WriteString("Current validator set:\n")
	for i, v := range r.Data {
		msg.WriteString(fmt.Sprintf("% 3d. %s", i, v))
		if i != len(r.Data)-1 {
			msg.WriteString("\n")
		}
	}

	return msg.Bytes(), nil
}
