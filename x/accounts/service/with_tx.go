package service

import "database/sql"

func (s *accountsService) WithTx(tx *sql.Tx) AccountsService {
	return &accountsService{
		dao: s.dao.WithTx(tx),
		db:  s.db,
	}
}
