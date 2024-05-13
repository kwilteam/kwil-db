package parse

import parseTypes "github.com/kwilteam/kwil-db/parse/types"

// typingVisitor performs type validation
type typingVisitor struct {
	// sqlCtx is the context of the current sql statement.
	// If we are not in a SQL statement, it is nil.
	sqlCtx *sqlContext
	// procCtx is the context of the current procedure.
	// If we are not in a procedure, it is nil.
	procCtx *procedureContext
	// errs is used for passing errors back to the caller.
	errs parseTypes.NativeErrorListener
}

// sqlContext is the context of the current SQL statement
type sqlContext struct{}

// procedureContext is the context of the current procedure.
type procedureContext struct{}
