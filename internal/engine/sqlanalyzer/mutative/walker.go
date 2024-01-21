package mutative

import "github.com/kwilteam/kwil-db/parse/sql/tree"

func NewMutativityWalker() *MutativityWalker {
	return &MutativityWalker{
		AstWalker: tree.NewBaseWalker(),
		Mutative:  false,
	}
}

type MutativityWalker struct {
	Mutative bool
	tree.AstWalker
}

func (m *MutativityWalker) EnterDeleteStmt(node *tree.DeleteStmt) error {
	m.Mutative = true
	return nil
}

func (m *MutativityWalker) EnterInsertStmt(node *tree.InsertStmt) error {
	m.Mutative = true
	return nil
}

func (m *MutativityWalker) EnterUpdateStmt(node *tree.UpdateStmt) error {
	m.Mutative = true
	return nil
}
