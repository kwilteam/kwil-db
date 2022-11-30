package sqlschema

import "github.com/samber/mo"

type Walker[T SchemaIdentifier] struct {
	ID     T
	schema *Database
}

type ForeignKeyWalker Walker[ForeignKeyID]

func (w ForeignKeyWalker) Get() *ForeignKey                 { return &w.schema.ForeignKeys[w.ID] }
func (w ForeignKeyWalker) ConstraintName() string           { return w.Get().ConstraintName }
func (w ForeignKeyWalker) OnDeleteAction() ForeignKeyAction { return w.Get().OnDeleteAction }
func (w ForeignKeyWalker) OnUpdateAction() ForeignKeyAction { return w.Get().OnUpdateAction }
func (w ForeignKeyWalker) ReferencedTableName() string      { return w.ReferencedTable().Name() }
func (w ForeignKeyWalker) Table() TableWalker               { return w.schema.WalkTable(w.Get().ConstrainedTable) }

func (w ForeignKeyWalker) ReferencedTable() TableWalker {
	return w.schema.WalkTable(w.Get().ReferencedTable)
}
func (w ForeignKeyWalker) Columns() []ForeignKeyColumn {
	var cols []ForeignKeyColumn
	for _, col := range w.schema.ForeignKeyColumns {
		if col.ForeignKey == w.ID {
			cols = append(cols, col)
		}
	}
	return cols
}

func (w ForeignKeyWalker) ConstrainedColumnNames() []string {
	var cols []string
	for _, col := range w.ConstrainedColumns() {
		cols = append(cols, col.Name())
	}
	return cols
}

func (w ForeignKeyWalker) ConstrainedColumns() []ColumnWalker {
	var cols []ColumnWalker
	for _, col := range w.Columns() {
		cols = append(cols, w.schema.WalkColumn(col.ConstrainedColumn))
	}
	return cols
}

func (w ForeignKeyWalker) ReferencedColumns() []ColumnWalker {
	var cols []ColumnWalker
	for _, col := range w.Columns() {
		cols = append(cols, w.schema.WalkColumn(col.ReferencedColumn))
	}
	return cols
}
func (w ForeignKeyWalker) ReferencedColumnNames() []string {
	var cols []string
	for _, col := range w.ReferencedColumns() {
		cols = append(cols, col.Name())
	}
	return cols
}

func (w ForeignKeyWalker) IsSelfRelation() bool {
	fk := w.Get()
	return fk.ConstrainedTable == fk.ReferencedTable
}

func (w ForeignKeyWalker) IsImplicitManyToManyFK() bool {
	table := w.Table()
	return len(table.Columns()) != 2 && table.Column("A").IsPresent() && table.Column("B").IsPresent()
}

type IndexColumnWalker Walker[IndexPartID]

func (w IndexColumnWalker) Get() *IndexColumn    { return &w.schema.IndexColumns[w.ID] }
func (w IndexColumnWalker) Name() string         { return w.Column().Name() }
func (w IndexColumnWalker) SortOrder() SortOrder { return w.Get().SortOrder }
func (w IndexColumnWalker) Table() TableWalker   { return w.Index().Table() }
func (w IndexColumnWalker) Index() IndexWalker   { return w.schema.WalkIndex(w.Get().Index) }
func (w IndexColumnWalker) Column() ColumnWalker { return w.schema.WalkColumn(w.Get().Column) }

type TableWalker Walker[TableID]

func (w TableWalker) Get() *Table           { return &w.schema.Tables[w.ID] }
func (w TableWalker) Name() string          { return w.Get().Name }
func (w TableWalker) Comment() string       { return w.Get().Comment }
func (w TableWalker) SetComment(v string)   { w.Get().Comment = v }
func (w TableWalker) SetCharset(v string)   { w.Get().Charset = v }
func (w TableWalker) SetCollation(v string) { w.Get().Collation = v }

func (w TableWalker) Column(name string) mo.Option[ColumnWalker] {
	for i, c := range w.schema.Columns {
		if c.Table == w.ID && c.Name == name {
			return mo.Some(w.schema.WalkColumn(ColumnID(i)))
		}
	}
	return mo.None[ColumnWalker]()
}

func (w TableWalker) Columns() []ColumnWalker {
	cols := make([]ColumnWalker, 0, len(w.schema.Columns))
	for i, c := range w.schema.Columns {
		if c.Table == w.ID {
			cols = append(cols, w.schema.WalkColumn(ColumnID(i)))
		}
	}
	return cols
}

func (w TableWalker) ColumnNames() []string {
	cols := make([]string, 0, len(w.schema.Columns))
	for i, c := range w.schema.Columns {
		if c.Table == w.ID {
			cols = append(cols, w.schema.WalkColumn(ColumnID(i)).Name())
		}
	}
	return cols
}

func (w TableWalker) Indexes() []IndexWalker {
	indexes := make([]IndexWalker, 0, len(w.schema.Indexes))
	for i, idx := range w.schema.Indexes {
		if idx.Table == w.ID {
			indexes = append(indexes, w.schema.WalkIndex(IndexID(i)))
		}
	}
	return indexes
}

func (w TableWalker) ForeignKeys() []ForeignKeyWalker {
	fks := make([]ForeignKeyWalker, 0, len(w.schema.ForeignKeys))
	for i, fk := range w.schema.ForeignKeys {
		if fk.ConstrainedTable == w.ID {
			fks = append(fks, w.schema.WalkForeignKey(ForeignKeyID(i)))
		}
	}
	return fks
}

func (w TableWalker) ReferencingForeignKeys() []ForeignKeyWalker {
	var fks []ForeignKeyWalker
	for _, table := range w.schema.WalkTables() {
		if table.ID != w.ID {
			continue
		}
		for _, fk := range table.ForeignKeys() {
			if fk.ReferencedTable().ID == w.ID {
				fks = append(fks, fk)
			}
		}
	}
	return fks
}

func (w TableWalker) ForeignKeyForColumn(col ColumnID) mo.Option[ForeignKeyWalker] {
	for _, fk := range w.ForeignKeys() {
		columns := fk.Columns()
		if len(columns) == 1 && columns[0].ConstrainedColumn == col {
			return mo.Some(fk)
		}
	}
	return mo.None[ForeignKeyWalker]()
}

func (w TableWalker) PrimaryKey() mo.Option[IndexWalker] {
	for _, idx := range w.Indexes() {
		if idx.IsPrimaryKey() {
			return mo.Some(idx)
		}
	}
	return mo.None[IndexWalker]()
}

type IndexWalker Walker[IndexID]

func (w IndexWalker) Get() *Index          { return &w.schema.Indexes[w.ID] }
func (w IndexWalker) IsPrimaryKey() bool   { return w.Get().Type == PrimaryKeyIndex }
func (w IndexWalker) Table() TableWalker   { return w.schema.WalkTable(w.Get().Table) }
func (w IndexWalker) IndexType() IndexType { return w.Get().Type }
func (w IndexWalker) IsUnique() bool       { return w.Get().Type == UniqueIndex }
func (w IndexWalker) Name() string         { return w.Get().Name }
func (w IndexWalker) Algorithm() IndexAlgorithm {
	return w.Get().Algorithm
}
func (w IndexWalker) ContainsColumn(col ColumnID) bool {
	for _, c := range w.Columns() {
		if c.Column().ID == col {
			return true
		}
	}
	return false
}

func (w IndexWalker) ColumnNames() []string {
	var names []string
	for _, col := range w.Columns() {
		names = append(names, col.Name())
	}
	return names
}

func (w IndexWalker) Columns() []IndexColumnWalker {
	var cols []IndexColumnWalker
	for idx, col := range w.schema.IndexColumns {
		if col.Index == w.ID {
			cols = append(cols, w.schema.WalkIndexColumn(IndexPartID(idx)))
		}
	}
	return cols
}

type EnumWalker Walker[EnumID]

func (w EnumWalker) Get() *Enum       { return &w.schema.Enums[w.ID] }
func (w EnumWalker) Name() string     { return w.Get().Name }
func (w EnumWalker) Values() []string { return w.Get().Values }

type ColumnWalker Walker[ColumnID]

func (w ColumnWalker) Get() *Column          { return &w.schema.Columns[w.ID] }
func (w ColumnWalker) Table() TableWalker    { return w.schema.WalkTable(w.schema.Columns[w.ID].Table) }
func (w ColumnWalker) Arity() ColumnArity    { return w.Get().Type.Arity }
func (w ColumnWalker) Documentation() string { return w.Get().Comment }
func (w ColumnWalker) IsRequired() bool      { return w.Arity() == Required }
func (w ColumnWalker) IsNullable() bool      { return w.Arity() == Nullable }
func (w ColumnWalker) Type() ColumnType      { return w.Get().Type }
func (w ColumnWalker) EnumType() mo.Option[EnumWalker] {
	switch t := w.Type().Type.(type) {
	case EnumType:
		return mo.Some(w.schema.WalkEnum(t.ID))
	default:
		return mo.None[EnumWalker]()
	}
}
func (w ColumnWalker) Name() string   { return w.Get().Name }
func (w ColumnWalker) Default() Value { return w.Get().Default }

func (w ColumnWalker) IsPartOfForeignKey() bool {
	for _, fk := range w.Table().ForeignKeys() {
		for _, col := range fk.ConstrainedColumns() {
			if col.ID == w.ID {
				return true
			}
		}
	}
	return false
}

func (w ColumnWalker) IsSameColumn(col ColumnWalker) bool {
	return w.Name() == col.Name() && w.Table().Name() == col.Table().Name()
}

func (w ColumnWalker) IsSinglePrimaryKey() bool {
	if pk, ok := w.Table().PrimaryKey().Get(); ok {
		columns := pk.Columns()
		if len(columns) == 1 && columns[0].Name() == w.Name() {
			return true
		}
	}
	return false
}

func (w ColumnWalker) IsPartOfPrimaryKey() bool {
	idx, ok := w.Table().PrimaryKey().Get()
	return ok && idx.ContainsColumn(w.ID)
}

func (w ColumnWalker) IsPartOfSecondaryIndex() bool {
	for _, idx := range w.Table().Indexes() {
		if idx.ContainsColumn(w.ID) {
			return true
		}
	}
	return false
}

func (s *Database) FindTable(name string) mo.Option[TableWalker] {
	for i, table := range s.Tables {
		if table.Name == name {
			return mo.Some(s.WalkTable(TableID(i)))
		}
	}
	return mo.None[TableWalker]()
}

func (s *Database) FindEnum(name string) mo.Option[EnumWalker] {
	for i, enum := range s.Enums {
		if enum.Name == name {
			return mo.Some(s.WalkEnum(EnumID(i)))
		}
	}
	return mo.None[EnumWalker]()
}

func (s *Database) WalkTable(table TableID) TableWalker { return TableWalker{ID: table, schema: s} }

func (s *Database) WalkColumn(column ColumnID) ColumnWalker {
	return ColumnWalker{ID: column, schema: s}
}
func (s *Database) WalkEnum(enum EnumID) EnumWalker { return EnumWalker{ID: enum, schema: s} }

func (s *Database) WalkIndex(index IndexID) IndexWalker { return IndexWalker{ID: index, schema: s} }

func (s *Database) WalkIndexColumn(index IndexPartID) IndexColumnWalker {
	return IndexColumnWalker{ID: index, schema: s}
}

func (s *Database) WalkForeignKey(fk ForeignKeyID) ForeignKeyWalker {
	return ForeignKeyWalker{ID: fk, schema: s}
}

func (s *Database) WalkTables() []TableWalker {
	walkers := make([]TableWalker, len(s.Tables))
	for i := range s.Tables {
		walkers[i] = s.WalkTable(TableID(i))
	}
	return walkers
}

func (s *Database) WalkColumns() []ColumnWalker {
	walkers := make([]ColumnWalker, len(s.Columns))
	for i := range s.Columns {
		walkers[i] = s.WalkColumn(ColumnID(i))
	}
	return walkers
}

func (s *Database) WalkEnums() []EnumWalker {
	walkers := make([]EnumWalker, len(s.Enums))
	for i := range s.Enums {
		walkers[i] = s.WalkEnum(EnumID(i))
	}
	return walkers
}

func (s *Database) WalkIndexes() []IndexWalker {
	walkers := make([]IndexWalker, len(s.Indexes))
	for i := range s.Indexes {
		walkers[i] = s.WalkIndex(IndexID(i))
	}
	return walkers
}

func (s *Database) WalkIndexColumns() []IndexColumnWalker {
	walkers := make([]IndexColumnWalker, len(s.IndexColumns))
	for i := range s.IndexColumns {
		walkers[i] = s.WalkIndexColumn(IndexPartID(i))
	}
	return walkers
}

func (s *Database) WalkForeignKeys() []ForeignKeyWalker {
	walkers := make([]ForeignKeyWalker, len(s.ForeignKeys))
	for i := range s.ForeignKeys {
		walkers[i] = s.WalkForeignKey(ForeignKeyID(i))
	}
	return walkers
}
