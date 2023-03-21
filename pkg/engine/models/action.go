package models

type Action struct {
	Name      string   `json:"name"`
	Public    bool     `json:"public"`
	Inputs    []string `json:"inputs"`
	Statement string   `json:"statement"`
}
