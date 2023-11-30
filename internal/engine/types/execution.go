package types

// ExecutionData is contextual data that is passed to a procedure during call / execution.
// It is scoped to the lifetime of a single execution.
type ExecutionData struct {
	// Dataset is the DBID of the dataset that was called.
	// Even if a procedure in another dataset is called, this will always be the original dataset.
	Dataset string

	// Procedure is the original procedure that was called.
	// Even if a nested procedure is called, this will always be the original procedure.
	Procedure string

	// Mutative indicates whether the execution can mutate state.
	Mutative bool

	// Args are the arguments that were passed to the procedure.
	Args []any

	// Caller is the binary identifier of the sender of the transaction.
	// This can be a public key, address, etc.
	Caller []byte

	// CallerIdentifier is a string identifier for the caller.
	// It is injected as a variable for usage in the query, under
	// the variable name "@caller".
	CallerIdentifier string
}
