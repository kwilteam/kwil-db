package dto

type Action struct {
	Name       string   `json:"name"`
	Inputs     []string `json:"inputs"`
	Public     bool     `json:"public"`
	Statements []string `json:"statements"`
}

func (a *Action) Clean() error {
	return runCleans(
		cleanIdent(&a.Name),
		cleanActionParameters(&a.Inputs),
	)
}
