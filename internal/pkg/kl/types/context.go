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
