package sqlschema

import (
	"ksl"
)

type Database struct {
	Name              string
	Tables            []Table
	Enums             []Enum
	Columns           []Column
	ForeignKeys       []ForeignKey
	ForeignKeyColumns []ForeignKeyColumn
	Indexes           []Index
	IndexColumns      []IndexColumn
	Extensions        []Extension
}

func NewDatabase(name string) Database {
	return Database{Name: name}
}

func (s *Database) AddColumn(column Column) ColumnID {
	columnID := ColumnID(len(s.Columns))
	s.Columns = append(s.Columns, column)
	return columnID
}

func (s *Database) AddEnum(name string, values ...string) EnumID {
	enumID := EnumID(len(s.Enums))
	s.Enums = append(s.Enums, Enum{Name: name})
	return enumID
}

func (s *Database) AddEnumVariant(e EnumID, variant string) {
	s.Enums[e].Values = append(s.Enums[e].Values, variant)
}

func (s *Database) AddIndex(index Index) IndexID {
	indexID := IndexID(len(s.Indexes))
	if index.Algorithm == UnspecifiedAlgo {
		index.Algorithm = BTreeAlgo
	}
	s.Indexes = append(s.Indexes, index)
	return indexID
}

func (s *Database) AddPrimaryKey(index Index) IndexID {
	indexID := IndexID(len(s.Indexes))
	index.Type = PrimaryKeyIndex
	if index.Algorithm == UnspecifiedAlgo {
		index.Algorithm = BTreeAlgo
	}
	s.Indexes = append(s.Indexes, index)
	return indexID
}

func (s *Database) AddUniqueIndex(index Index) IndexID {
	indexID := IndexID(len(s.Indexes))
	index.Type = UniqueIndex
	if index.Algorithm == UnspecifiedAlgo {
		index.Algorithm = BTreeAlgo
	}
	s.Indexes = append(s.Indexes, index)
	return indexID
}

func (s *Database) AddIndexColumn(col IndexColumn) IndexPartID {
	indexColumnID := IndexPartID(len(s.IndexColumns))
	s.IndexColumns = append(s.IndexColumns, col)
	return indexColumnID
}

func (s *Database) AddForeignKey(fk ForeignKey) ForeignKeyID {
	foreignKeyID := ForeignKeyID(len(s.ForeignKeys))
	s.ForeignKeys = append(s.ForeignKeys, fk)
	return foreignKeyID
}

func (s *Database) AddForeignKeyColumn(fkc ForeignKeyColumn) {
	s.ForeignKeyColumns = append(s.ForeignKeyColumns, fkc)
}

func (s *Database) AddTable(table Table) TableID {
	tableID := TableID(len(s.Tables))
	s.Tables = append(s.Tables, table)
	return tableID
}

func (s *Database) AddExtension(ext Extension) ExtensionID {
	extID := ExtensionID(len(s.Extensions))
	s.Extensions = append(s.Extensions, ext)
	return extID
}

func (s Database) TableNames() []string {
	names := make([]string, len(s.Tables))
	for i, table := range s.Tables {
		names[i] = table.Name
	}
	return names
}

type Table struct {
	Name      string
	Comment   string
	Charset   string
	Collation string
}

type Column struct {
	Table         TableID
	Name          string
	Type          ColumnType
	Default       Value
	AutoIncrement bool
	Comment       string
	Charset       string
	Collation     string
}

type ColumnType struct {
	Type  ksl.Type
	Raw   string
	Arity ColumnArity
}

type Index struct {
	Table     TableID
	Name      string
	Type      IndexType
	Algorithm IndexAlgorithm
}

type IndexColumn struct {
	Index     IndexID
	Column    ColumnID
	SortOrder SortOrder
}

type Procedure struct {
	Name       string
	Definition string
}

type Enum struct {
	Name   string
	Values []string
}

type View struct {
	Name       string
	Definition string
}

type ForeignKey struct {
	ConstrainedTable TableID
	ReferencedTable  TableID
	ConstraintName   string
	OnDeleteAction   ForeignKeyAction
	OnUpdateAction   ForeignKeyAction
}

type ForeignKeyColumn struct {
	ForeignKey        ForeignKeyID
	ConstrainedColumn ColumnID
	ReferencedColumn  ColumnID
}
