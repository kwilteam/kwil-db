package metadata

import (
	"database/sql"
	"fmt"
	"kwil/x"
	"sync"
	"time"
)

var expired_deadline = x.NewDeadline(time.Duration(1) * time.Millisecond)

type Connector interface {
	GetConnectionInfo(wallet string) (string, error)
}

type ConnectorFunc func(string) (string, error)

func (fn ConnectorFunc) GetConnectionInfo(wallet string) (string, error) {
	return fn(wallet)
}

func LocalConnectionInfo(wallet string) (string, error) {
	return fmt.Sprintf("postgres://localhost:5432/%s?sslmode=disable", wallet), nil
}

func NewProvider(db *sql.DB, closeDbConnection bool) *ConnectionProvider {
	return &ConnectionProvider{
		db:                db,
		connections:       make(map[string]string),
		mu:                &sync.Mutex{},
		refresh:           expired_deadline,
		closeDbConnection: closeDbConnection,
	}
}

type ConnectionProvider struct {
	db                *sql.DB
	closeDbConnection bool
	connections       map[string]string
	mu                *sync.Mutex
	refresh           *x.Deadline
}

func (m *ConnectionProvider) Close() error {
	if !m.closeDbConnection {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.db.Close()
}

func (m *ConnectionProvider) GetConnectionInfo(wallet string) (string, error) {
	return "postgres://postgres:postgres@localhost:5432/kwil?sslmode=disable", nil
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.refresh.HasExpired() {
		m.connections = make(map[string]string)
		rows, err := m.db.Query("select wallet, db_connection_url as url from wallet_info")
		if err != nil {
			return "", err
		} else {
			for rows.Next() {
				var wallet string
				var url string
				err = rows.Scan(&wallet, &url)
				if err != nil {
					return "", err
				}

				m.connections[wallet] = url
			}
		}
	}

	return m.connections[wallet], nil
}
