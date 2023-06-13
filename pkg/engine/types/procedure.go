package types

type Procedure struct {
	Name       string   `json:"name"`
	Args       []string `json:"inputs"`
	Public     bool     `json:"public"`
	Statements []string `json:"statements"`
}

func (p *Procedure) Clean() error {
	return runCleans(
		cleanIdent(&p.Name),
		cleanActionParameters(&p.Args),
	)
}

func (p *Procedure) Identifier() string {
	return p.Name
}
