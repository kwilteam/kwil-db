package types

type Procedure struct {
	Name       string     `json:"name"`
	Args       []string   `json:"inputs"`
	Public     bool       `json:"public"`
	Modifiers  []Modifier `json:"modifiers"`
	Statements []string   `json:"statements"`
}

// Clean cleans the procedure, and returns an error if it is invalid.
func (p *Procedure) Clean() error {
	for _, m := range p.Modifiers {
		if err := m.Clean(); err != nil {
			return err
		}
	}

	return runCleans(
		cleanIdent(&p.Name),
		cleanActionParameters(&p.Args),
	)
}

// IsMutative returns true if the procedure is mutative.
func (p *Procedure) IsMutative() bool {
	for _, m := range p.Modifiers {
		if m == ModifierView {
			return false
		}
	}

	return true
}

func (p *Procedure) RequiresAuthentication() bool {
	for _, m := range p.Modifiers {
		if m == ModifierAuthenticated {
			return true
		}
	}

	return false
}

func (p *Procedure) IsOwnerOnly() bool {
	for _, m := range p.Modifiers {
		if m == ModifierOwner {
			return true
		}
	}

	return false
}

func (p *Procedure) EnsureContainsModifier(mod Modifier) {
	contains := false
	for _, m := range p.Modifiers {
		if m == mod {
			contains = true
		}
	}

	if !contains {
		p.Modifiers = append(p.Modifiers, mod)
	}
}
