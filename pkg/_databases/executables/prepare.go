package executables

// the preparer is used to generate statements and prepare inputs.
// it handles applying attributes and modifiers, checks that data types are correct, and
// ensures that required inputs are provided.
type preparer struct {
	executable *executable

	// the user inputs that are being used, mapped by name
	inputs map[string][]byte

	caller string
}
