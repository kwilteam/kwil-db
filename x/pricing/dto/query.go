package dto

import spec "kwil/x/sqlx"

type Query struct {
	Database string  `json:"database"`
	Owner    string  `json:"owner"`
	Caller   string  `json:"caller"`
	Query    string  `json:"query"`
	Inputs   []Input `json:"inputs"`
}

// TODO: this is the same as the one in x/sqlx/models
type Input struct {
	Position int           `json:"position"`
	Type     spec.DataType `json:"type"`
	Value    string        `json:"value"`
}
