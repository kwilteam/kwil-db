package db

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/serialize"
)

/*
	Typesit could be worth doing a migration instead of upgrading on the fly
*/

func decodeVersionedProcedures(meta []*VersionedMetadata) ([]*types.Procedure, error) {
	procedures := make([]*types.Procedure, len(meta))
	for i, proc := range meta {
		procedure, err := decodeProcedure(proc.Version, proc.Data)
		if err != nil {
			return nil, err
		}

		procedures[i] = procedure
	}

	return procedures, nil
}

func decodeProcedure(version uint, procedureBytes []byte) (*types.Procedure, error) {
	procedure := &types.Procedure{}
	err := serialize.DecodeInto(procedureBytes, procedure)
	if err != nil {
		return nil, err
	}

	for {
		switch version {
		case 1:
			procedure = upgradeProcedure_v1_To_v2(procedure)
			version++
		case procedureVersion:
			return procedure, nil
		default:
			return nil, fmt.Errorf("unknown procedure version %d", version)
		}
	}
}

// upgradeProcedure_v1_To_v2 upgrades a procedure from version 1 to version
// If a v1 procedure is private, we change it to public and add an owner modifier.
func upgradeProcedure_v1_To_v2(oldProc *types.Procedure) *types.Procedure {
	newProc := &types.Procedure{
		Name:       oldProc.Name,
		Args:       oldProc.Args,
		Statements: oldProc.Statements,
		Public:     oldProc.Public,
		Modifiers:  oldProc.Modifiers,
	}

	if !oldProc.Public {
		newProc.EnsureContainsModifier(types.ModifierOwner)
		newProc.Public = true
	}

	return newProc
}
