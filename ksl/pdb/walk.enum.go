package pdb

import (
	"ksl/syntax/ast"
)

type EnumWalker struct {
	db *Db
	id EnumID
}

func (w EnumWalker) Db() *Db              { return w.db }
func (w EnumWalker) AstEnum() *ast.Enum   { return w.db.Ast.GetEnum(w.id) }
func (w EnumWalker) Name() string         { return w.AstEnum().GetName() }
func (w EnumWalker) ID() EnumID           { return w.id }
func (w EnumWalker) Get() EnumAnnotations { return w.db.Types.EnumAnnotations[w.id] }
func (w EnumWalker) DatabaseName() string {
	data := w.Get()
	name := data.MappedName
	if name == "" {
		name = w.Name()
	}
	return name
}

func (w EnumWalker) Values() []EnumValueWalker {
	enum := w.AstEnum()
	values := make([]EnumValueWalker, len(enum.Values))
	for i := range enum.Values {
		values[i] = EnumValueWalker{db: w.db, id: MakeEnumValueID(w.id, IndexID(i))}
	}
	return values
}

type EnumValueWalker struct {
	db *Db
	id EnumValueID
}

func (w EnumValueWalker) Db() *Db          { return w.db }
func (w EnumValueWalker) ID() EnumValueID  { return w.id }
func (w EnumValueWalker) Enum() EnumWalker { return EnumWalker{db: w.db, id: w.id.Enum()} }
func (w EnumValueWalker) Documentation() string {
	return w.Enum().AstEnum().Values[w.id.Value()].Documentation()
}
func (w EnumValueWalker) Name() string { return w.Enum().AstEnum().Values[w.id.Value()].GetName() }
func (w EnumValueWalker) DatabaseName() string {
	info := w.db.Types.EnumAnnotations[w.id.Enum()]
	name := info.MappedValues[w.id]
	if name == "" {
		name = w.Name()
	}
	return name
}
