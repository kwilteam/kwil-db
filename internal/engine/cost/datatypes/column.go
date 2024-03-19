package datatypes

type ColumnDef struct {
	Relation *TableRef
	Name     string
}

func ColumnUnqualified(name string) *ColumnDef {
	return &ColumnDef{Name: name}
}

func Column(table *TableRef, name string) *ColumnDef {
	return &ColumnDef{Relation: table, Name: name}
}
