package tree

// PGMarshaler is a procedural language AST or AST node that can be marshaled into a string.
type PGMarshaler interface {
	MarshalPG(info *SystemInfo) (string, error)
}

// Clause is the core building block of a procedural script.
// Anything that can be written to the procedural language is a clause.
// Clauses should always end with a semicolon.
type Clause interface {
	PGMarshaler
	clause() // private method to ensure that only clauses can be assigned to a Clause
}
