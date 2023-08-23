package queue

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrQueueFull = errors.New("engine queue is full")
)

func New(size int) *Queue {
	q := &Queue{
		queueChannel: make(chan *struct{}, size),
		waiters:      make(map[*struct{}]chan struct{}),
		running:      make(chan struct{}),
	}

	go q.run()

	return q
}

/*
Queue is a queue that allows callers to wait in the queue, and only pops a new
one off when the previous one has been processed.

A better name for this might be "ordered mutex".
*/
type Queue struct {
	// queueChannel is a buffered channel that allows callers to wait in the queue
	queueChannel chan *struct{}

	// waiters is a map of callers that are waiting in the queue
	waiters map[*struct{}]chan struct{}

	// running is a channel for the current caller to signal that they are done processing
	running chan struct{}

	// mutex is used to protect the queueChannel
	// before adding, callers will check if its full
	// its possible to have a race condition where there is 1 spot left in the queue
	// and 2 callers try to add at the same time, resulting in one blocking
	// and therefore not being in order
	mu sync.Mutex
}

// addToQueue adds a caller to the queue
// it is returned a channel that will be closed when it is their turn to process
func (q *Queue) addToQueue() (<-chan struct{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// check if channel is full
	if len(q.queueChannel) == cap(q.queueChannel) {
		return nil, ErrQueueFull
	}

	ptr := &struct{}{}

	// add caller to queue
	q.queueChannel <- ptr

	// return channel that will be closed when it is their turn to process
	retChan := make(chan struct{})

	q.waiters[ptr] = retChan

	return retChan, nil
}

func (q *Queue) Wait(ctx context.Context) (func(), error) {
	retChan, err := q.addToQueue()
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-retChan:
		return func() {
			// signal that we are done processing
			q.running <- struct{}{}
		}, nil
	}
}

func (q *Queue) run() {
	for {
		// get next caller
		ptr := <-q.queueChannel

		// retrieve caller from map, remove from map
		callerChan, ok := q.waiters[ptr]
		delete(q.waiters, ptr)

		// if caller is not in map, continue
		if !ok {
			continue
		}

		// close channel to signal that it is their turn to process
		close(callerChan)

		// wait for caller to be done processing
		<-q.running
	}
}
