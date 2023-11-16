package sqlite

import (
	"sync"

	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/kwil-db/core/utils/random"
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// Savepoint creates a new savepoint.
func (c *Connection) Savepoint() (sql.Savepoint, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isReadonly() {
		return nil, ErrReadOnlyConn
	}

	// if a savepoint is the outer-most, then we should checkpoint the WAL when it is closed
	shouldCheckpoint := c.activeSavepoint.CompareAndSwap(false, true)

	saveName := "tx_" + randomSavepointName(8)

	err := execute(c.conn, "SAVEPOINT "+saveName)
	if err != nil {
		return nil, err
	}

	return &Savepoint{
		conn:     c.conn,
		saveName: saveName,
		closeFn: func() error {
			defer func() {
				c.activeSavepoint.Store(false)
			}()
			if shouldCheckpoint {
				err := c.checkpointWal()
				if err != nil {
					return err
				}
			}
			return nil
		},
	}, nil
}

// Savepoint is a checkpoint in the state of the database.
// It can be rolled back to, or committed.
type Savepoint struct {
	mu       sync.Mutex
	conn     *sqlite.Conn
	saveName string

	closed bool

	// closeFn is called when the savepoint is either committed or rolled back.
	closeFn func() error
}

// With both Commit and Rollback, if the checkpoint fails, it doesn't matter

// Commit commits the savepoint and releases it
func (sp *Savepoint) Commit() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.closed {
		return nil
	}
	sp.closed = true

	err := execute(sp.conn, "RELEASE "+sp.saveName)
	if err != nil {
		return err
	}

	return sp.closeFn()
}

// Rollback rolls back the savepoint and releases it
func (sp *Savepoint) Rollback() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.closed {
		return nil
	}
	sp.closed = true

	err := execute(sp.conn, "ROLLBACK TO "+sp.saveName)
	if err != nil {
		return err
	}

	err = execute(sp.conn, "RELEASE "+sp.saveName)
	if err != nil {
		return err
	}

	return sp.closeFn()
}

func randomSavepointName(length int) string {
	if length < 2 {
		panic("Length must be at least 2 to generate a valid savepoint name.")
	}
	return random.String(length)
}
