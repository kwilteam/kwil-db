package types

import (
	"testing"

	"github.com/kwilteam/kwil-db/knode/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgDatabaseWrite_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDatabaseWrite
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDatabaseWrite{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgDatabaseWrite{
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
