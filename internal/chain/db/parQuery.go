package db

import (
	"encoding/json"

	"github.com/kwilteam/kwil-db/pkg/types/db"
)

type StoredPQ struct {
	Query  *string         `json:"query"`
	Params *[]db.Parameter `json:"params"`
}

func StoreParQuer(pq *db.ParameterizedQuery, d KVBasic) error {
	pqKey := getPQKey(pq.Name)
	storePQ := StoredPQ{
		Query:  &pq.Query,
		Params: &pq.Parameters,
	}

	b, err := storePQ.Bytes()
	if err != nil {
		return err
	}

	err = d.Set(pqKey, b)
	if err != nil {
		return err
	}
	return nil
}

func getPQKey(n string) []byte {
	return append([]byte("pq"), []byte(n)...)
}

func (s *StoredPQ) Bytes() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

func (s *StoredPQ) Unmarshal(b []byte) error {
	return json.Unmarshal(b, s)
}

func (s *StoredPQ) toParamaterizedQuery(n string) *db.ParameterizedQuery {
	return &db.ParameterizedQuery{
		Name:       n,
		Query:      *s.Query,
		Parameters: *s.Params,
	}
}

func (d *DB) GetParQuer(n string) (*db.ParameterizedQuery, error) {
	pqKey := getPQKey(n)
	b, err := d.Get(pqKey)
	if err != nil {
		return nil, err
	}
	pq := &StoredPQ{}
	err = pq.Unmarshal(b)
	if err != nil {
		return nil, err
	}
	return pq.toParamaterizedQuery(n), nil
}
