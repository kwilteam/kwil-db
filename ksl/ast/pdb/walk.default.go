package pdb

import (
	"ksl/syntax/nodes"
)

type DefaultValueWalker struct {
	model ModelID
	field FieldID
	db    *Db
	attr  *DefaultAnnotation
}

func (w DefaultValueWalker) Db() *Db { return w.db }
func (w DefaultValueWalker) AstAnnotation() *nodes.Annotation {
	return w.db.Ast.GetAnnotation(w.attr.SourceAnnot)
}
func (w DefaultValueWalker) Value() nodes.Expression {
	return w.attr.Value
}

func (w DefaultValueWalker) Field() ScalarFieldWalker {
	scalar := w.db.Types.ScalarFields[MakeModelFieldID(w.model, w.field)]
	return ScalarFieldWalker{w.model, w.field, w.db, scalar}
}
