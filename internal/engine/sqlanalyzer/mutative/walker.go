package mutative

import "github.com/kwilteam/kwil-db/parse/sql/tree"

func NewMutativityWalker() *MutativityWalker {
	return &MutativityWalker{
		AstListener: tree.NewBaseListener(),
		Mutative:    false,
	}
}

type MutativityWalker struct {
	Mutative bool
	tree.AstListener
}

func (m *MutativityWalker) EnterDeleteCore(node *tree.DeleteCore) error {
	m.Mutative = true
	return nil
}

func (m *MutativityWalker) EnterInsertCore(node *tree.InsertCore) error {
	m.Mutative = true
	return nil
}

func (m *MutativityWalker) EnterUpdateCore(node *tree.UpdateCore) error {
	m.Mutative = true
	return nil
}
