package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgDatabaseWrite{}, "kwil/DatabaseWrite", nil)
	cdc.RegisterConcrete(&MsgCreateDatabase{}, "kwil/CreateDatabase", nil)
	cdc.RegisterConcrete(&MsgDDL{}, "kwil/DDL", nil)
	cdc.RegisterConcrete(&MsgDefineQuery{}, "kwil/DefineQuery", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDatabaseWrite{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateDatabase{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDDL{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDefineQuery{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
