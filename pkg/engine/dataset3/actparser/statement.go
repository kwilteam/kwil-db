package actparser

type StmtType string

const (
	StmtTypeDML           StmtType = "dml"
	StmtTypeExtensionCall StmtType = "extension_call"
	StmtTypeActionCall    StmtType = "action_call"
)

type ActionStmt interface {
	StmtType() StmtType
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
	Method string
	Args   []string
}

type DMLStmt struct {
	Statement string
}

func (s *ExtensionCallStmt) StmtType() StmtType {
	return StmtTypeExtensionCall
}

func (s *ActionCallStmt) StmtType() StmtType {
	return StmtTypeActionCall
}

func (s *DMLStmt) StmtType() StmtType {
	return StmtTypeDML
}
