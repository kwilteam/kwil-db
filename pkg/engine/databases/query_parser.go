package databases

type queryAst interface {
	ToSQL() (string, error)
}

type queryParser interface {
	Parse(string) (queryAst, error)
}
