package types

type Extension struct {
	Name           string            `json:"name"`
	Initialization map[string]string `json:"initialization"`
	Alias          string            `json:"alias"`
}

func (e *Extension) Clean() error {
	return runCleans(
		cleanIdent(&e.Name),
		cleanIdent(&e.Alias),
	)
}
