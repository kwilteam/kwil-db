package types

type Procedure struct {
	Name       string   `json:"name"`
	Args       []string `json:"args"`
	Public     bool     `json:"public"`
	Statements []string `json:"statements"`
}

func (p *Procedure) Identifier() string {
	return p.Name
}
