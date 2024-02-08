package types

type Procedure struct {
	Name        string     `json:"name"`
	Annotations []string   `json:"annotations,omitempty"`
	Args        []string   `json:"inputs"`
	Public      bool       `json:"public"`
	Modifiers   []Modifier `json:"modifiers"`
	Statements  []string   `json:"statements"`
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
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

// IsView returns true if the procedure has a view modifier.
func (p *Procedure) IsView() bool {
	for _, m := range p.Modifiers {
		if m == ModifierView {
			return true
		}
	}

	return false
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
