package driver

import (
	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/go-sqlite/sqlitex"
)

type Transaction struct {
	conn     *sqlite.Conn
	saveName string
}

func (tx *Transaction) Execute(sql string, args ...interface{}) error {
	return sqlitex.Execute(tx.conn, sql, &sqlitex.ExecOptions{
		Args: args,
	})
}

// With both Commit and Rollback, if the checkpoint fails, it doesn't matter

func (tx *Transaction) Commit() error {
	defer tx.checkpointWal() // it doesn't matter if this fails
	return sqlitex.Execute(tx.conn, "RELEASE "+tx.saveName, nil)
}

func (tx *Transaction) Rollback() error {
	defer tx.checkpointWal() // it doesn't matter if this fails
	err := sqlitex.Execute(tx.conn, "ROLLBACK TO "+tx.saveName, nil)
	if err != nil {
		return err
	}

	return sqlitex.Execute(tx.conn, "RELEASE "+tx.saveName, nil)
}

func (tx *Transaction) checkpointWal() error {
	return sqlitex.Execute(tx.conn, "PRAGMA wal_checkpoint(TRUNCATE);", nil)
}

func Begin(conn *Connection) (*Transaction, error) {
	saveName := "tx_" + randomSavepointName(8)
	err := sqlitex.Execute(conn.Conn, "SAVEPOINT "+saveName, nil)
	if err != nil {
		return nil, err
	}
	return &Transaction{conn: conn.Conn, saveName: saveName}, nil
}

func (c *Connection) Begin() (*Transaction, error) {
	return Begin(c)
}
