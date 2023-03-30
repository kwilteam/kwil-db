package models

type Action struct {
	Name       string   `json:"name" clean:"lower"`
	Public     bool     `json:"public"`
	Inputs     []string `json:"inputs" clean:"lower"`
	Statements []string `json:"statements"`
}
