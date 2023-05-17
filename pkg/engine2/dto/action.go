package dto

type Action struct {
	Name       string   `json:"name" clean:"lower"`
	Inputs     []string `json:"inputs" clean:"lower"`
	Public     bool     `json:"public"`
	Statements []string `json:"statements"`
}
