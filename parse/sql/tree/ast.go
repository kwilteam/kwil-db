package tree

type Ast interface {
	ToSQL() (string, error)
	Accepter
}
