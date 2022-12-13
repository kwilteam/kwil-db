package models

import types "kwil/x/sqlx"

type Attribute struct {
	Type  string `json:"type"`
	Value any    `json:"value,omitempty"`
}

func (a *Attribute) Validate(c *Column) error {
	// check if attribute type is valid and convert it to the correct type
	attr, err := types.Conversion.ConvertAttribute(a.Type)
	if err != nil {
		return err
	}

	// check if attribute value is valid: e.g. if it is a MIN or MAX attribute, the value must be an int
	return types.Validation.CorrectAttributeType(a.Value, attr, c.Type)
}
