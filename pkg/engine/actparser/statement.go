package actparser

type ActionStmt interface {
	StmtType() string
}

//type CallStmt struct {
//	Method    string
//	Args      []string
//	Receivers []string
//}

type ExtensionCallStmt struct {
	Extension string
	Method    string
	Args      []string
	Receivers []string
}

type ActionCallStmt struct {
	Database  string
	Method    string
	Args      []string
	Receivers []string
}

type DMLStmt struct {
	Statement string
}

func (s *ExtensionCallStmt) StmtType() string {
	return "extension_call"
}

func (s *ActionCallStmt) StmtType() string {
	return "action_call"
}

func (s *DMLStmt) StmtType() string {
	return "dml"
}
