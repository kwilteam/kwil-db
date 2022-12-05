package schema

type QueryType int

const (
	Create QueryType = iota
	Update
	Delete
)

type ColumnMap map[string]any

type ColumnValues map[string]any

type PreparedStatement interface {
	ToSql(rows ...ColumnValues) string
}
