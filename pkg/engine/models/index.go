package models

type Index struct {
	Name    string   `json:"name" clean:"lower"`
	Table   string   `json:"table" clean:"lower"`
	Columns []string `json:"columns" clean:"lower"`
}
