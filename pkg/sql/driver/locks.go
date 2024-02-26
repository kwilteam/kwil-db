package driver

import (
	"sync"
	"time"
)

// LockTypes are meant to emulate SQLite's locking types, with the exception of
// Kwil's UNLOCKED type, which is also used to indicate that there is no
// connection.
// https://www.sqlite.org/lockingv3.html
type LockType uint8

var (
	locks = make(map[string]*sync.Mutex) // true if locked, false if unlocked
)

const (
	LockWaitTimeMs = 100
)

func acquireLock(dbid string, timeout time.Duration) error {
	return nil
	//if _, ok := locks[dbid]; !ok {
	//	locks[dbid] = &sync.Mutex{}
	//}
	//
	//lockAcquired := make(chan bool)
	//mu := locks[dbid]
	//
	//// start a goroutine that will try to acquire the lock
	//go func() {
	//	mu.Lock()
	//	lockAcquired <- true
	//}()
	//
	//select {
	//case <-lockAcquired:
	//	return nil
	//case <-time.After(timeout):
	//	return ErrLockWaitTimeout
	//}
}

func releaseLock(dbid string) {
	if _, ok := locks[dbid]; !ok {
		return
	}

	locks[dbid].Unlock()
}

const (
	// UNLOCKED is the default lock type
	// if there is no connection, it is unlocked
	LOCK_TYPE_UNLOCKED LockType = iota

	LOCK_TYPE_READ_ONLY
	LOCK_TYPE_READ_WRITE
)

// Readable returns true if the connection is readable
func (c *Connection) Readable() bool {
	return c.lock == LOCK_TYPE_READ_ONLY || c.lock == LOCK_TYPE_READ_WRITE
}

// Writable returns true if the connection is writable
func (c *Connection) Writable() bool {
	return c.lock == LOCK_TYPE_READ_WRITE
}
