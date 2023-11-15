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

	listExample = `$ kwild validator list
Current validator set:
  0. {pubkey = 9fed4b19eab0cf87370bff0c7ef04cfcff5b268d096578d3ef5ae3c1010939d8, power = 1}
  1. {pubkey = f1693096d252eac5e366436314996ec39c189dffb997ad7414ed306e3d9244c4, power = 1}
  2. {pubkey = f28e3cd6d11e9c5d2d00ea8b9c1e18cc01e568f0b186888fb274567461774fbc, power = 1}`
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
