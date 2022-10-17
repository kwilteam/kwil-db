package events

import (
	"fmt"
	"sync"
)

/* the queue is a FIFO queue. It is used to store block heights that are received from the websocket
and are waiting to be confirmed.  The queue is a slice of int64s, which are the block numbers
*/

type Queue struct {
	queue []int64
	head  int64
	tail  int64
	len   uint16
	mu    sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		queue: make([]int64, 0),
		head:  0,
		tail:  0,
		len:   0,
	}
}

func (q *Queue) Append(height int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, height)
	q.tail = height
	if q.len == 0 {
		q.head = height
	}
	q.len++
}

func (q *Queue) Pop() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.len == 0 {
		return -1
	}
	ret := q.queue[0]
	q.queue = q.queue[1:]
	q.head = ret
	q.len--
	return ret
}

func (q *Queue) Len() uint16 {
	q.mu.Lock()
	l := q.len
	q.mu.Unlock()
	return l
}

func (q *Queue) Head() int64 {
	q.mu.Lock()
	h := q.head
	q.mu.Unlock()
	return h
}

func (q *Queue) Tail() int64 {
	q.mu.Lock()
	t := q.tail
	q.mu.Unlock()
	return t
}

func (q *Queue) Print() {
	q.mu.Lock()
	for _, v := range q.queue {
		fmt.Println(v)
	}
	q.mu.Unlock()
}
