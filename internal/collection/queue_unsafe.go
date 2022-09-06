package collection

type queueUnsafeImpl[T any] struct {
	el *[]T
}

func newQueueUnsafeImpl[T any]() *queueUnsafeImpl[T] {
	return &queueUnsafeImpl[T]{new([]T)}
}

func (q *queueUnsafeImpl[T]) IsEmpty() bool {
	return len(*q.el) == 0
}

func (q *queueUnsafeImpl[T]) Add(e T) {
	a := append(*q.el, e)
	q.el = &a
}

func (q *queueUnsafeImpl[T]) AddIf(e T, fn func(e T) bool) bool {
	if !fn(e) {
		return false
	}

	q.Add(e)

	return true
}

func (q *queueUnsafeImpl[T]) Poll() (v T, ok bool) {
	if q.IsEmpty() {
		return
	}

	e := (*q.el)[0]
	s := (*q.el)[1:]
	q.el = &s

	return e, true
}

func (q *queueUnsafeImpl[T]) Size() int {
	return len(*q.el)
}

func (q *queueUnsafeImpl[T]) isSynchronized() bool {
	return false
}
