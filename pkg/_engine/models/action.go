package models

type Action struct {
	Name       string   `json:"name" clean:"lower"`
	Public     bool     `json:"public"`
	Inputs     []string `json:"inputs"`
	Statements []string `json:"statements"`
}
