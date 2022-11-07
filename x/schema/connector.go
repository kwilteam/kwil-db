package schema

import "fmt"

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
