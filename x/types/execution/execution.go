package execution

type ExecutionBody struct {
	Database string // the id
	Query    string
	Caller   string
	Inputs   []*UserInput
}
