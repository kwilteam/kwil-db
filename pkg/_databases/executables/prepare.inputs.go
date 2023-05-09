package executables

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
)

// paramOrWhere is used to share code between parameters and wheres
type paramOrWhere interface {
	GetName() string
	GetColumn() string
	GetStatic() bool
	GetModifier() spec.ModifierType
	GetValue() *spec.KwilAny
}

// prepareInput prepares either a parameter or a where, applying attributes and modifiers,
// as well as ensuring that the data type is correct and that all required inputs are provided.
func (p *preparer) prepareInput(param paramOrWhere) (*spec.KwilAny, error) {
	// start out with the default value
	val := param.GetValue()

	// if not static, get the user input
	if !param.GetStatic() {
		byteVal, ok := p.inputs[param.GetName()]
		if !ok {
			return nil, fmt.Errorf(`required parameter "%s" was not provided`, param.GetName())
		}

		newVal, err := spec.NewFromSerial(byteVal)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse parameter "%s": %w`, param.GetName(), err)
		}

		val = newVal
	}

	column, ok := p.executable.Columns[param.GetColumn()]
	if !ok {
		// this should never happen
		// the column should be validated when the query is created
		return nil, fmt.Errorf(`column "%s" could not be found in the DBI.  this is a server issue`, param.GetColumn())
	}

	// apply any modifiers
	if err := p.applyModifier(val, param.GetModifier()); err != nil {
		return nil, fmt.Errorf(`failed to apply modifier to parameter "%s": %w`, param.GetName(), err)
	}

	// apply any attributes
	if err := p.applyAttributes(column, val); err != nil {
		return nil, fmt.Errorf(`failed to apply attributes to parameter "%s": %w`, param.GetName(), err)
	}

	// validate the data type.  This must come last in case of caller modifier
	if column.Type != val.Type() {
		t1 := val.Type()
		t2 := column.Type
		return nil, fmt.Errorf(`parameter "%s" is of type "%s" but should be of type "%s"`, param.GetName(), t1.String(), t2.String())
	}

	return val, nil
}

func (p *preparer) applyAttributes(col *databases.Column[*spec.KwilAny], val *spec.KwilAny) error {
	for _, attr := range col.Attributes {
		fn := spec.AttributeFuncs[attr.Type]
		if fn == nil {
			// this should never happen
			return fmt.Errorf(`attribute "%d" is not supported`, attr.Type)
		}

		var err error
		val, err = fn(val, attr.Value)
		if err != nil {
			return fmt.Errorf(`failed to apply attribute "%d" to column "%s": %w`, attr.Type, col.Name, err)
		}
	}

	return nil
}

func (p *preparer) applyModifier(val *spec.KwilAny, m spec.ModifierType) error {
	switch m {
	case spec.CALLER:
		newVal, err := spec.NewExplicit(p.caller, spec.STRING)
		if err != nil {
			return err
		}

		*val = *newVal
	}

	return nil
}
