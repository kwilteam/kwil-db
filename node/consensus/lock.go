package consensus

import "sync"

type PriorityLockQueue struct {
	mtx    sync.Mutex
	active bool
	queue  []chan struct{}
}

type queueFunc func(q []chan struct{}, c chan struct{}) []chan struct{}

func appendTo[E any](q []E, c E) []E {
	return append(q, c)
}

func prependTo[E any](q []E, c E) []E {
	if cap(q) > len(q) { // with extra capacity, shift in-place to avoid realloc
		q = append(q, c) // extend, allowing runtime to over-allocate
		copy(q[1:], q)   // shift right
		q[0] = c         // insert at front
	} else {
		q = append([]E{c}, q...)
	}
	return q
}

func (pl *PriorityLockQueue) lock(qf queueFunc) {
	pl.mtx.Lock()
	if !pl.active {
		pl.active = true
		pl.mtx.Unlock()
		return
	}

	ch := make(chan struct{})
	pl.queue = qf(pl.queue, ch)
	pl.mtx.Unlock()

	<-ch // wait
}

func (pl *PriorityLockQueue) Lock() {
	pl.lock(appendTo) // back of the line
}

func (pl *PriorityLockQueue) PriorityLock() {
	pl.lock(prependTo) // jump the line
	// NOTE: this is intended for only one serial caller to PriorityLock, like
	// commit(), not multiple. If there is another PriorityLock caller before
	// the first unblocks, the second one will take the front of the line.
}

func (pl *PriorityLockQueue) Unlock() {
	pl.mtx.Lock()

	if len(pl.queue) == 0 {
		pl.active = false
		pl.mtx.Unlock()
		return
	}

	// Wake up the next in line
	ch := pl.queue[0]
	pl.queue = pl.queue[1:]
	// pl.active = true

	pl.mtx.Unlock()

	close(ch)
}
