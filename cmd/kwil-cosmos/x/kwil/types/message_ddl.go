package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kwilteam/kwil-db/internal/utils"
)

const TypeMsgDDL = "ddl"

var _ sdk.Msg = &MsgDDL{}

func NewMsgDDL(creator string, dbid string, ddl string) *MsgDDL {
	return &MsgDDL{
		Creator: creator,
		Dbid:    dbid,
		Ddl:     ddl,
	}
}

func (msg *MsgDDL) Route() string {
	return RouterKey
}

func (msg *MsgDDL) Type() string {
	return TypeMsgDDL
}

func (msg *MsgDDL) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	utils.PanicIfError(err)
	return []sdk.AccAddress{creator}
}

func (msg *MsgDDL) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDDL) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
