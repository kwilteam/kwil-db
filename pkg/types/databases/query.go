package databases

import (
	"bytes"
	"encoding/gob"
	"kwil/pkg/execution"
	"kwil/pkg/types/data_types/any_type"
)

type SQLQuery[T anytype.AnyValue] struct {
	Name   string              `json:"name" clean:"lower"`
	Type   execution.QueryType `json:"type" clean:"is_enum,query_type"`
	Table  string              `json:"table" clean:"lower"`
	Params []*Parameter[T]     `json:"parameters,omitempty" clean:"struct"`
	Where  []*WhereClause[T]   `json:"where,omitempty" clean:"struct"`
}

func (s *SQLQuery[T]) EncodeGOB() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := gob.NewEncoder(buf).Encode(s)
	return buf.Bytes(), err
}

func (s *SQLQuery[T]) DecodeGOB(b []byte) error {
	var qry SQLQuery[T]
	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(&qry)
	if err != nil {
		return err
	}
	*s = qry
	return nil
}

// thanks for nothing generics
func (s *SQLQuery[T]) ListParamColumns() []string {
	var cols []string
	for _, p := range s.Params {
		cols = append(cols, p.Column)
	}
	return cols
}

func (s *SQLQuery[T]) ListParamColumnsAsAny() []any {
	var cols []interface{}
	for _, p := range s.Params {
		cols = append(cols, p.Column)
	}
	return cols
}

func (s *SQLQuery[T]) ListWhereColumns() []string {
	var cols []string
	for _, w := range s.Where {
		cols = append(cols, w.Column)
	}
	return cols
}

func (s *SQLQuery[T]) ListWhereColumnsAsAny() []any {
	var cols []interface{}
	for _, w := range s.Where {
		cols = append(cols, w.Column)
	}
	return cols
}
