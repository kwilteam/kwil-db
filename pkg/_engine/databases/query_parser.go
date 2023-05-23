package databases

type queryAst interface {
	ToSQL() (string, error)
}

type queryParser interface {
	Parse(string) (queryAst, error)
}

var parser queryParser

func init() {
	// TODO: add parser
}

type SqlParseFunc func(string) (queryAst, error)

type DbidFunc func(owner string, name string) string
