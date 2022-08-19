package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kwilteam/kwil-db/internal/utils"
)

const TypeMsgCreateDatabase = "create_database"

var _ sdk.Msg = &MsgCreateDatabase{}

func NewMsgCreateDatabase(creator string, seed string) *MsgCreateDatabase {
	return &MsgCreateDatabase{
		Creator: creator,
		Seed:    seed,
	}
}

func (msg *MsgCreateDatabase) Route() string {
	return RouterKey
}

func (msg *MsgCreateDatabase) Type() string {
	return TypeMsgCreateDatabase
}

func (msg *MsgCreateDatabase) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	utils.PanicIfError(err)
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateDatabase) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateDatabase) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
