package sql_parser

type ActionContext map[string]any
type TableContext struct {
	Columns      []string
	Indexes      []string // index names
	IndexColumns []string // columns(parameters) of index, corresponding to index name
	PrimaryKeys  []string
}

type DatabaseContext struct {
	Tables  map[string]TableContext
	Actions map[string]ActionContext
}

func NewTableContext() TableContext {
	return TableContext{
		Columns:      []string{},
		Indexes:      []string{},
		PrimaryKeys:  []string{},
		IndexColumns: []string{},
	}
}

func NewActionContext() ActionContext {
	return ActionContext{}
}

func NewDatabaseContext() DatabaseContext {
	return DatabaseContext{
		Tables:  map[string]TableContext{},
		Actions: map[string]ActionContext{},
	}
}
