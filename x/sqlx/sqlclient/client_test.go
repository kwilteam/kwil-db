package sqlclient_test

import (
	"testing"

	"kwil/x/sqlx/sqlclient"
)

func Test_OpenClient(t *testing.T) {
	client, err := sqlclient.Open("postgres://postgres:postgres@localhost:5432/kwil?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Ping(); err != nil {
		t.Fatal(err)
	}
}

func Test_CreateConnectionString(t *testing.T) {
	conn := sqlclient.CreateConnectionString(
		&sqlclient.ConnectionInfo{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			Database: "kwil",
			SSL:      "disable",
		},
	)
	if conn != "postgres://postgres:postgres@localhost:5432/kwil?sslmode=disable" {
		t.Fatal("Connection string is not correct")
	}
}
