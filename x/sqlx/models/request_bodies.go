package models

import "strings"

type PayloadBody uint8

const (
	INVALID_MESSAGE PayloadBody = iota
	CREATE_DATABASE
	EDIT_DATABASE
	DROP_DATABASE
	EXECUTABLE
)

func (p PayloadBody) Byte() byte {
	return byte(p)
}

type QueryTx struct {
	Query    string       `json:"query" yaml:"query"`
	Inputs   []*UserInput `json:"inputs" yaml:"inputs"`
	Database string       `json:"database" yaml:"database"`
	Owner    string       `json:"owner" yaml:"owner"`
}

func (q *QueryTx) GetSchemaName() string {
	return strings.ToLower(q.Database + "_" + q.Owner)
}

type DropDatabase struct {
	Name  string `json:"name" yaml:"name"`
	Owner string `json:"owner" yaml:"owner"`
}

type CreateDatabase struct {
	Database []byte `json:"database" yaml:"database"`
}
