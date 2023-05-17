package builder

import "github.com/kwilteam/kwil-db/pkg/engine/tree"

type insert struct {
	ins *tree.InsertStmt
}

func newInsert(insertType tree.InsertType) *insert {
	return &insert{
		ins: &tree.InsertStmt{
			InsertType: insertType,
			Columns:    make([]string, 0),
			Values:     make([][]tree.Expression, 0),
		},
	}
}

type InsertBuilderWithAlias interface {
	AliasSelector[InsertBuilder]
	InsertBuilder
}

type InsertBuilder interface {
	Columns(...string) InsertBuilder
	ValuePicker
}

type ValuePicker interface {
	Values([]tree.Expression) ValuePicker
}

func (i *insert) Table(name string) InsertBuilderWithAlias {
	i.ins.Table = name
	return i
}

func (i *insert) As(alias string) InsertBuilder {
	i.ins.TableAlias = alias
	return i
}

func (i *insert) Columns(columns ...string) InsertBuilder {
	i.ins.Columns = columns
	return i
}

func (i *insert) Values(values []tree.Expression) ValuePicker {
	i.ins.Values = append(i.ins.Values, values)
	return i
}

func (i *insert) Upsert(upsert *tree.Upsert) InsertBuilder {
	i.ins.Upsert = upsert
	return i
}
