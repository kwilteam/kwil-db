package serialize

// a procedure is a collection of operations that can be executed as a single unit
// it is atomic, and has local variables
type procedure_v2 struct {
	Name    string      `json:"name"`
	Args    []string    `json:"args"`
	Scoping uint8       `json:"scoping"`
	Body    []Operation `json:"body"`
}
