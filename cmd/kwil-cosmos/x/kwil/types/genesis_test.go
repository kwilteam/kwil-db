package types_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{

				DatabasesList: []types.Databases{
					{
						Index: "0",
					},
					{
						Index: "1",
					},
				},
				DdlList: []types.Ddl{
					{
						Index: "0",
					},
					{
						Index: "1",
					},
				},
				DdlindexList: []types.Ddlindex{
					{
						Index: "0",
					},
					{
						Index: "1",
					},
				},
				QueryidsList: []types.Queryids{
					{
						Index: "0",
					},
					{
						Index: "1",
					},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated databases",
			genState: &types.GenesisState{
				DatabasesList: []types.Databases{
					{
						Index: "0",
					},
					{
						Index: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated ddl",
			genState: &types.GenesisState{
				DdlList: []types.Ddl{
					{
						Index: "0",
					},
					{
						Index: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated ddlindex",
			genState: &types.GenesisState{
				DdlindexList: []types.Ddlindex{
					{
						Index: "0",
					},
					{
						Index: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated queryids",
			genState: &types.GenesisState{
				QueryidsList: []types.Queryids{
					{
						Index: "0",
					},
					{
						Index: "0",
					},
				},
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
