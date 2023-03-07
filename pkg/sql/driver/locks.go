package driver

// LockTypes are meant to emulate SQLite's locking types, with the exception of
// Kwil's UNLOCKED type, which is also used to indicate that there is no
// connection.
// https://www.sqlite.org/lockingv3.html
type LockType uint8

const (
	// UNLOCKED is the default lock type
	// if there is no connection, it is unlocked
	LOCK_TYPE_UNLOCKED LockType = iota

	LOCK_TYPED_SHARED
	LOCK_TYPE_RESERVED
	LOCK_TYPE_PENDING
	LOCK_TYPE_EXCLUSIVE
)

// Readable returns true if the connection is readable
func (c *Connection) Readable() bool {
	return c.lock == LOCK_TYPED_SHARED || c.lock == LOCK_TYPE_RESERVED || c.lock == LOCK_TYPE_PENDING || c.lock == LOCK_TYPE_EXCLUSIVE
}

// Writable returns true if the connection is writable
func (c *Connection) Writable() bool {
	return c.lock == LOCK_TYPE_EXCLUSIVE
}
