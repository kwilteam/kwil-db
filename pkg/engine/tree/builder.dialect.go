package tree

import "github.com/doug-martin/goqu/v9"

var Builder = &builder{}

type builder struct {
	dialect goqu.DialectWrapper
}

func init() {
	Builder = &builder{
		// TODO: there are some custom dialect modifications we need to make, Brennan will do later
		dialect: goqu.Dialect("sqlite3"),
	}
}
