package sqlmigrate

import "ksl/sqlschema"

type SqlFlavor interface {
}

type SqlDiffFlavor interface {
	ColumnTypeChange(prev, next sqlschema.ColumnWalker) ColumnTypeChange
}
