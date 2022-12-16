package repository

import (
	"kwil/x/sqlx/sqlclient"

	"github.com/doug-martin/goqu/v9"
)

type DAO interface {
	NewDepositQuery() DepositQuery
}

type dao struct{}

var DB *sqlclient.DB

func pgQB() *goqu.Database {
	return goqu.Dialect("postgres").DB(DB.DB)
}

func NewDAO() DAO {
	return &dao{}
}

func NewDB(db *sqlclient.DB) {
	DB = db
}

func (d *dao) NewDepositQuery() DepositQuery {
	return &depositQuery{}
}
