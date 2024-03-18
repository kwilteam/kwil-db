package tree

import types "github.com/kwilteam/kwil-db/internal/engine/procedural/types"

// SystemInfo holds information about the entire database system.
// It is used for quick lookup of system information.
type SystemInfo struct {
	// Schemas are the schemas in the system.
	Schemas map[string]*SchemaInfo

	Context *ProcedureContext
}

// SchemaInfo holds information about a postgres schema.
type SchemaInfo struct {
	// Types are the types in the schema.
	Types map[string]*types.CompositeTypeDefinition
	// Procedures are the procedures in the schema.
	Procedures map[string]*types.Procedure
}

// ProcedureContext contains information about the procedure call.
type ProcedureContext struct {
	Variables map[string]types.DataType // all known variables
}
