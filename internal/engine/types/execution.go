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

	// Signer is the address of public key that signed the incoming transaction.
	Signer []byte

	// Caller is a string identifier for the signer.
	// It is derived from the signer's registered authenticator.
	// It is injected as a variable for usage in the query, under
	// the variable name "@caller".
	Caller string
}

func (e *ExecutionData) Clean() error {
	return runCleans(
		cleanDBID(&e.Dataset),
		cleanIdent(&e.Procedure),
	)
}
