package sqlite

import (
	"github.com/kwilteam/kwil-db/pkg/utils/random"
)

func (c *Connection) Savepoint() (*Savepoint, error) {
	return beginSavepoint(c)
}

func beginSavepoint(c *Connection) (*Savepoint, error) {
	saveName := "tx_" + randomSavepointName(8)

	err := c.Execute("SAVEPOINT " + saveName)
	if err != nil {
		return nil, err
	}
	return &Savepoint{conn: c, saveName: saveName}, nil
}

type Savepoint struct {
	conn     *Connection
	saveName string
}

// With both Commit and Rollback, if the checkpoint fails, it doesn't matter

// Commit commits the savepoint and releases it
func (sp *Savepoint) Commit() error {
	return sp.conn.Execute("RELEASE " + sp.saveName)
}

// CommitAndCheckpoint commits the savepoint, releases it, and checkpoints the WAL
func (sp *Savepoint) CommitAndCheckpoint() error {
	err := sp.Commit()
	if err != nil {
		return err
	}

	return sp.conn.CheckpointWal()
}

// Rollback rolls back the savepoint and releases it
func (sp *Savepoint) Rollback() error {
	err := sp.conn.Execute("ROLLBACK TO " + sp.saveName)
	if err != nil {
		return err
	}

	return sp.conn.Execute("RELEASE " + sp.saveName)
}

func randomSavepointName(length int) string {
	if length < 2 {
		panic("Length must be at least 2 to generate a valid savepoint name.")
	}
	return random.String(length)
}
