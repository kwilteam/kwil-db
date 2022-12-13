package models

import (
	"fmt"
	types "kwil/x/sqlx"
)

type WhereClause struct {
	Column   string `json:"column"`
	Static   bool   `json:"static"`
	Operator string `json:"operator,omitempty"`
	Value    any    `json:"value,omitempty"`
	Modifier string `json:"modifier,omitempty"`
}

func (w *WhereClause) Validate(table *Table) error {
	// check if column exists
	col := table.GetColumn(w.Column)
	if col == nil {
		return fmt.Errorf(`column "%s" does not exist`, w.Column)
	}

	mod, err := types.Conversion.ConvertModifier(w.Modifier)
	if err != nil {
		return fmt.Errorf(`invalid modifier for where clause on column "%s": %w`, w.Column, err)
	}

	// validate operator
	_, err = types.Conversion.ConvertComparisonOperator(w.Operator)
	if err != nil {
		return fmt.Errorf(`invalid operator for where clause on column "%s": %w`, w.Column, err)
	}

	if w.Static {

		// check if value is set
		if w.Value == nil {
			return fmt.Errorf(`value must be set for non-fillable where clause on column "%s"`, w.Column)
		}

		// check the default value type matches the column type
		err = types.Validation.CompareAnyToKwilString(w.Value, col.Type)
		if err != nil {
			return fmt.Errorf(`value type "%s" does not match column type "%s" for where clause on column "%s"`, w.Value, col.Type, w.Column)
		}
	} else { // not static: users can fill in the value
		if w.Value != nil {
			return fmt.Errorf(`value must not be set for fillable where clause on column "%s"`, w.Column)
		}

		if mod == types.CALLER {
			return fmt.Errorf(`modifier must not be caller for fillable where clause on column "%s"`, w.Column)
		}
	}

	return nil
}

// including these getters to fulfill arger interface
func (w *WhereClause) getColumn() string {
	return w.Column
}

func (w *WhereClause) getModifier() string {
	return w.Modifier
}

func (w *WhereClause) getStatic() bool {
	return w.Static
}

func (w *WhereClause) getValue() any {
	return w.Value
}

func (w *WhereClause) buildArg(tbl *Table, position int) (*Arg, error) {
	return buildArg(tbl, position, w)
}
