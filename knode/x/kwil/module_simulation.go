package kwil

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	kwilsimulation "github.com/kwilteam/kwil-db/knode/internal/x/kwil/simulation"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
	"github.com/kwilteam/kwil-db/knode/testutil/sample"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = kwilsimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgDatabaseWrite = "op_weight_msg_database_write"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDatabaseWrite int = 100

	opWeightMsgCreateDatabase = "op_weight_msg_create_database"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateDatabase int = 100

	opWeightMsgDDL = "op_weight_msg_ddl"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDDL int = 100

	opWeightMsgDefineQuery = "op_weight_msg_define_query"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDefineQuery int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	kwilGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&kwilGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgDatabaseWrite int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDatabaseWrite, &weightMsgDatabaseWrite, nil,
		func(_ *rand.Rand) {
			weightMsgDatabaseWrite = defaultWeightMsgDatabaseWrite
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDatabaseWrite,
		kwilsimulation.SimulateMsgDatabaseWrite(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgCreateDatabase int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateDatabase, &weightMsgCreateDatabase, nil,
		func(_ *rand.Rand) {
			weightMsgCreateDatabase = defaultWeightMsgCreateDatabase
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateDatabase,
		kwilsimulation.SimulateMsgCreateDatabase(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDDL int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDDL, &weightMsgDDL, nil,
		func(_ *rand.Rand) {
			weightMsgDDL = defaultWeightMsgDDL
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDDL,
		kwilsimulation.SimulateMsgDDL(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDefineQuery int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDefineQuery, &weightMsgDefineQuery, nil,
		func(_ *rand.Rand) {
			weightMsgDefineQuery = defaultWeightMsgDefineQuery
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDefineQuery,
		kwilsimulation.SimulateMsgDefineQuery(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
