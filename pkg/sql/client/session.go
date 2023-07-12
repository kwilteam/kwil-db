package client

import (
	"github.com/kwilteam/kwil-db/pkg/sql"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type Session struct {
	ses *sqlite.Session
}

func (c *SqliteClient) BeginSession() (sql.Session, error) {
	ses, err := c.conn.CreateSession()
	if err != nil {
		return nil, err
	}

	return &Session{
		ses: ses,
	}, nil
}

func (c *Session) Delete() error {
	return c.ses.Delete()
}

func (c *Session) GenerateChangeset() ([]byte, error) {
	cs, err := c.ses.GenerateChangeset()
	if err != nil {
		return nil, err
	}

	return cs.Export(), nil
}
