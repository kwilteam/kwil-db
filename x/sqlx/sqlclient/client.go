package sqlclient

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(conn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type ConnectionInfo struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSL      string
}

const (
	DEF_HOST     = "localhost"
	DEF_PORT     = "5432"
	DEF_USER     = "postgres"
	DEF_PASSWORD = "postgres"
	DEF_DATABASE = "kwil"
	DEF_SSL      = "disable"
)

func CreateConnectionString(c *ConnectionInfo) string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.Database + "?sslmode=" + c.SSL
}
