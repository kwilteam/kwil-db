package executables

// An ExecutionBody is sent during execution
// TODO: @brennan I believe this belongs in spec / somewhere else, since the database field is used to determine which DBI to use
type ExecutionBody struct {
	Database string       `json:"database" yaml:"database" clean:"lower"` // the id
	Query    string       `json:"query" yaml:"query" clean:"lower"`       // the name of the query
	Inputs   []*UserInput `json:"inputs" yaml:"inputs"`                   // the inputs to the query
}

// UserInput is sent during execution
type UserInput struct {
	Name string `json:"name" yaml:"name" clean:"lower"` // Name is the name of the input

	Value []byte `json:"value" yaml:"value"` // Value is the value of the input
}
