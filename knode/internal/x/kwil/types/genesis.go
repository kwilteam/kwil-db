package types

import (
	"fmt"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		DatabasesList: []Databases{},
		DdlList:       []Ddl{},
		DdlindexList:  []Ddlindex{},
		QueryidsList:  []Queryids{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in databases
	databasesIndexMap := make(map[string]struct{})

	for _, elem := range gs.DatabasesList {
		index := string(DatabasesKey(elem.Index))
		if _, ok := databasesIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for databases")
		}
		databasesIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in ddl
	ddlIndexMap := make(map[string]struct{})

	for _, elem := range gs.DdlList {
		index := string(DdlKey(elem.Index))
		if _, ok := ddlIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for ddl")
		}
		ddlIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in ddlindex
	ddlindexIndexMap := make(map[string]struct{})

	for _, elem := range gs.DdlindexList {
		index := string(DdlindexKey(elem.Index))
		if _, ok := ddlindexIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for ddlindex")
		}
		ddlindexIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in queryids
	queryidsIndexMap := make(map[string]struct{})

	for _, elem := range gs.QueryidsList {
		index := string(QueryidsKey(elem.Index))
		if _, ok := queryidsIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for queryids")
		}
		queryidsIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
