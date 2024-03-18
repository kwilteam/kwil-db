package types

type Procedure struct {
	// Name is the name of the procedure.
	// It should always be lower case.
	Name string

	// Parameters are the parameters of the procedure.
	Parameters []*CompositeTypeField

	// Public is true if the procedure is public.
	Public bool

	// Returns is the return type of the procedure.
	Returns ProcedureReturn
}

// ProcedureReturn is what a procedure returns.
// It can be nil, a single value, or a table.
type ProcedureReturn interface {
	returns()
}

type ReturnTable struct {
	// Columns are the columns of the table.
	Columns []*CompositeTypeField
}

func (ReturnTable) returns() {}
