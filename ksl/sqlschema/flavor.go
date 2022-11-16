package sqlschema

type SqlFlavor interface {
}

type SqlDiffFlavor interface {
	ColumnTypeChange(prev, next ColumnWalker) ColumnTypeChange
}
