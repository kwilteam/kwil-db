package databases

import (
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

// DatasetContext is a context for a dataset.
// Once provided, it should not be modified.
type DatasetContext struct {
	// SqlParseFunc is a function that parses a sql string into a sqlite statement.
	SqlParseFunc SqlParseFunc

	// IdFunc is a function that generates a dbid from an owner and name.
	IdFunc DbidFunc

	// Name is the name of the dataset.
	Name string
	// Owner is the owner of the dataset.
	Owner string
}

// A database is a single deployed instance of kwil-db.
// It contains a SQLite file
type Dataset struct {
	Ctx     *DatasetContext
	conn    *sqlite.Connection
	actions map[string]*Action
}

func (d *Dataset) CreateAction() *ActionBuilder {
	return actionBuilder(d)
}
