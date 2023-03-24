package types

type ActionContext map[string]any
type TableContext struct {
	Columns     []string
	Indexes     []string
	PrimaryKeys []string
}

type DatabaseContext struct {
	Tables  map[string]TableContext
	Actions map[string]ActionContext
}

func NewTableContext() TableContext {
	return TableContext{
		Columns:     []string{},
		Indexes:     []string{},
		PrimaryKeys: []string{},
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
