package dto

import (
	"bytes"
	"encoding/gob"
	"kwil/x/execution"
)

type SQLQuery struct {
	Name   string              `json:"name"`
	Type   execution.QueryType `json:"type"`
	Table  string              `json:"table"`
	Params []*Parameter        `json:"parameters,omitempty"`
	Where  []*WhereClause      `json:"where,omitempty"`
}

func (s *SQLQuery) EncodeGOB() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := gob.NewEncoder(buf).Encode(s)
	return buf.Bytes(), err
}

func (s *SQLQuery) DecodeGOB(b []byte) error {
	var qry SQLQuery
	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(&qry)
	if err != nil {
		return err
	}
	*s = qry
	return nil
}

// thanks for nothing generics
func (s *SQLQuery) ListParamColumns() []string {
	var cols []string
	for _, p := range s.Params {
		cols = append(cols, p.Column)
	}
	return cols
}

func (s *SQLQuery) ListParamColumnsAsAny() []any {
	var cols []interface{}
	for _, p := range s.Params {
		cols = append(cols, p.Column)
	}
	return cols
}

func (s *SQLQuery) ListWhereColumns() []string {
	var cols []string
	for _, w := range s.Where {
		cols = append(cols, w.Column)
	}
	return cols
}

func (s *SQLQuery) ListWhereColumnsAsAny() []any {
	var cols []interface{}
	for _, w := range s.Where {
		cols = append(cols, w.Column)
	}
	return cols
}
