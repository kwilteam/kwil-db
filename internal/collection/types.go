package collection

type Queue[T any] interface {
	IsEmpty() bool
	Add(e T)
	AddIf(e T, fn func(v T) bool) bool
	Poll() (T, bool)
	Size() int
	isSynchronized() bool
}

func NewQueueUnsafe[T any]() Queue[T] {
	return newQueueUnsafeImpl[T]()
}

func NewQueue[T any]() Queue[T] {
	return newQueueSync[T]()
}

//goland:noinspection GoUnusedExportedFunction
func MakeSafeQueue[T any](queue *Queue[T]) *Queue[T] {
	q := *queue
	if q.isSynchronized() {
		return queue
	}

	var w Queue[T] = makeSafeQueue(queue)
	return &w
}
