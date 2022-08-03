package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDatabaseWrite = "database_write"

var _ sdk.Msg = &MsgDatabaseWrite{}

func NewMsgDatabaseWrite(creator string, database string, parQuer string, data string) *MsgDatabaseWrite {
	return &MsgDatabaseWrite{
		Creator:  creator,
		Database: database,
		ParQuer:  parQuer,
		Data:     data,
	}
}

func (msg *MsgDatabaseWrite) Route() string {
	return RouterKey
}

func (msg *MsgDatabaseWrite) Type() string {
	return TypeMsgDatabaseWrite
}

func (msg *MsgDatabaseWrite) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDatabaseWrite) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDatabaseWrite) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
