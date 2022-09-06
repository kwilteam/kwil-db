package collection

import (
	"sync"
)

type queueImpl[T any] struct {
	inner Queue[T]
	mu    sync.Mutex
}

func newQueueSync[T any]() *queueImpl[T] {
	return &queueImpl[T]{
		inner: newQueueUnsafeImpl[T](),
		mu:    sync.Mutex{},
	}
}

func makeSafeQueue[T any](queue *Queue[T]) *queueImpl[T] {
	q := *queue
	return &queueImpl[T]{
		inner: q,
		mu:    sync.Mutex{},
	}
}

func (q *queueImpl[T]) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.inner.IsEmpty()
}

func (q *queueImpl[T]) Add(e T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.inner.Add(e)
}

func (q *queueImpl[T]) AddIf(e T, fn func(e T) bool) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !fn(e) {
		return false
	}

	q.inner.Add(e)

	return true
}

func (q *queueImpl[T]) Poll() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.inner.Poll()
}

func (q *queueImpl[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.inner.Size()
}

func (q *queueImpl[T]) isSynchronized() bool {
	return true
}

//type queueNode[T any] struct {
//	element *T
//	next    *queueNode[T]
//}
//
//type queueLinkedList[T any] struct {
//	head *queueNode[T]
//	tail *queueNode[T]
//	mu   sync.Mutex
//}
//
//func (q *queueLinkedList[T]) IsEmpty() bool {
//	q.mu.Lock()
//	defer q.mu.Unlock()
//
//	return q.head == nil
//}
//
//func (q *queueLinkedList[T]) Add(element *T) {
//	q.mu.Lock()
//	defer q.mu.Unlock()
//
//	node := &queueNode[T]{element, nil}
//	if q.isEmptyUnsafe() {
//		q.head = node
//	} else {
//		q.tail.next = node
//	}
//	q.tail = node
//}
//
//func (q *queueLinkedList[T]) Poll() *T {
//	q.mu.Lock()
//	defer q.mu.Unlock()
//
//	if q.isEmptyUnsafe() {
//		return nil
//	}
//
//	tmp := q.head
//
//	q.head = q.head.next
//	if q.head == nil {
//		q.tail = nil
//	}
//
//	return tmp.element
//}
//
//func (q *queueLinkedList[T]) isEmptyUnsafe() bool {
//	return q.head == nil
//}
