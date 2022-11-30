package sqlclient

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4"
)

type Client struct {
	DB *sql.DB
}

func NewClient(conn string) *Client {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		panic(err)
	}
	return &Client{
		DB: db,
	}
}
