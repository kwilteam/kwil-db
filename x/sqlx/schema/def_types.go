package schema

type QueryType int

const (
	Create QueryType = iota
	Update
	Delete
)

type ColumnMap map[string]string

type ColumnValues map[string]any

type PreparedStatement interface {
	ToSql(rows ...ColumnValues) string
}

type preparedStatement struct {
	// The SQL statement
	Sql string
	// The inputs by order
	Parameters map[int]string    // mapping of parameter number to type
	Defaults   map[string]string // mapping of column name to default value
	InputOrder map[string]int    // maps column name to parameter number
}

type statementInput struct {
	name string
	tp   string
}
