package sqlschema

type Pair[T any] struct {
	Prev, Next T
}

func MakePair[T any](prev, next T) Pair[T] { return Pair[T]{prev, next} }
