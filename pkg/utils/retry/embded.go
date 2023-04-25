package retry

type embedder[T any] struct {
	Target  T
	Retrier Retrier[T]
}
