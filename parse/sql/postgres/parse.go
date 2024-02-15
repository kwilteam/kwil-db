package postgres

type CheckSyntaxFunc func(query string) error

var CheckSyntax CheckSyntaxFunc = doNothing

// doNothing is a placeholder for the CheckSyntaxFunc when cgo is disabled.
func doNothing(_ string) error {
	return nil
}
