package execution

type ExecutionBody struct {
	Database string       `json:"database" yaml:"database"` // the id
	Query    string       `json:"query" yaml:"query"`       // the name of the query
	Inputs   []*UserInput `json:"inputs" yaml:"inputs"`     // the inputs to the query
}
