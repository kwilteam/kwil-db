package types

import (
	"fmt"
	"strings"
)

type Modifier string

const (
	// View means that an action does not modify the database.
	ModifierView Modifier = "VIEW"

	// Authenticated requires that the caller is identified.
	ModifierAuthenticated Modifier = "AUTHENTICATED"
)

func (m *Modifier) IsValid() bool {
	upper := strings.ToUpper(m.String())

	return upper == ModifierView.String() ||
		upper == ModifierAuthenticated.String()
}

func (m *Modifier) Clean() error {
	if !m.IsValid() {
		return fmt.Errorf("invalid modifier: %s", m.String())
	}

	*m = Modifier(strings.ToUpper(m.String()))

	return nil
}

func (m Modifier) String() string {
	return string(m)
}
