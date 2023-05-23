package builder

type TableSelector[T any] interface {
	Table(string) T
}

type AliasSelector[T any] interface {
	As(string) T
}
