package events

import (
	"fmt"
	"math/big"
	"sync"
)

/* the queue is a FIFO queue. It is used to store headers that are received from the websocket
and are waiting to be confirmed.  The queue is a slice of big.Ints, which are the block numbers
*/

type Queue struct {
	queue []*big.Int
	head  *big.Int
	tail  *big.Int
	len   int
	mu    sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		queue: make([]*big.Int, 0),
		head:  big.NewInt(0),
		tail:  big.NewInt(0),
		len:   0,
	}
}

func (q *Queue) Append(height *big.Int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, height)
	q.tail = height
	if q.len == 0 {
		q.head = height
	}
	q.len++
}

func (q *Queue) Pop() *big.Int {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.len == 0 {
		return nil
	}
	ret := q.queue[0]
	q.queue = q.queue[1:]
	q.head = ret
	q.len--
	return ret
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.len
}

func (q *Queue) Head() *big.Int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.head
}

func (q *Queue) Tail() *big.Int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.tail
}

func (q *Queue) Print() {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, v := range q.queue {
		fmt.Println(v)
	}
}
