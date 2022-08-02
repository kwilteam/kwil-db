package types

import (
	"testing"

	"github.com/kwilteam/kwil-db/knode/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateDatabase_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateDatabase
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateDatabase{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateDatabase{
				Creator: sample.AccAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
