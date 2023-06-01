package sqlite

import sqlparser "github.com/kwilteam/kwil-db/pkg/sql_parser"

// TODO: import parsefunc here once Gavin has implemented it

func parseSql(sql string) (str string, err error) {
	// catch any potential panics
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	ast, err := sqlparser.Parse(sql)
	if err != nil {
		return "", err
	}

	return ast.ToSQL()
}
