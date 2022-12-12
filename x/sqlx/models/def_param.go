package models

import (
	"fmt"
	types "kwil/x/sqlx/spec"
)

type Param struct {
	Column   string `json:"column"`
	Static   bool   `json:"static"`
	Value    any    `json:"value,omitempty"`
	Modifier string `json:"modifier,omitempty"`
}

func (p *Param) Validate(table *Table) error {
	// check if column exists
	col := table.GetColumn(p.Column)
	if col == nil {
		return fmt.Errorf(`column "%s" does not exist`, p.Column)
	}

	mod, err := types.Conversion.ConvertModifier(p.Modifier)
	if err != nil {
		return fmt.Errorf(`invalid modifier for parameter on column "%s": "%s". `, p.Column, p.Modifier)
	}

	if p.Static {

		// check if value is set
		if p.Value == nil {
			return fmt.Errorf(`value must be set for non-fillable parameter on column "%s"`, p.Column)
		}

		err = types.Validation.CompareAnyToKwilString(p.Value, col.Type)

		// check if value type matches column type
		if err != nil {
			return fmt.Errorf(`value type "%s" does not match column type "%s" for parameter on column "%s"`, p.Value, col.Type, p.Column)
		}
	} else { // not static: users can fill in the value
		if p.Value != nil {
			return fmt.Errorf(`value must not be set for fillable parameter on column "%s"`, p.Column)
		}

		if mod == types.CALLER {
			return fmt.Errorf(`modifier must not be caller for fillable parameter on column "%s"`, p.Column)
		}
	}

	return nil
}
