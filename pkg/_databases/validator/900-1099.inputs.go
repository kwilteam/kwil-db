package validator

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
)

/*
###################################################################################################

	Inputs: 900-999

###################################################################################################
*/

// validateInputs validates both parameters and where clauses
func (v *Validator) validateInputs(params []*databases.Parameter[*spec.KwilAny], where []*databases.WhereClause[*spec.KwilAny], table *databases.Table[*spec.KwilAny]) error {
	if len(params) > MAX_PARAM_PER_QUERY {
		return violation(errorCode902, fmt.Errorf(`too many parameters: %v > %v`, len(params), MAX_PARAM_PER_QUERY))
	}

	if len(where) > MAX_WHERE_PER_QUERY {
		return violation(errorCode903, fmt.Errorf(`too many where clauses: %v > %v`, len(where), MAX_WHERE_PER_QUERY))
	}

	inputColumns := make(map[string]struct{}) // for guaranteeing that each column is only used at most once
	inputNames := make(map[string]struct{})   // for guaranteeing that each name is only used at most once for both params and where
	for _, param := range params {
		if _, ok := inputNames[param.Name]; ok {
			return violation(errorCode900, fmt.Errorf(`duplicate parameter/where-clause name "%s"`, param.Name))
		}
		inputNames[param.Name] = struct{}{}

		// checking column uniqueness.  this is only done for parameters, not where clauses
		if _, ok := inputColumns[param.Column]; ok {
			return violation(errorCode901, fmt.Errorf(`duplicate parameter column "%s"`, param.Column))
		}
		inputColumns[param.Column] = struct{}{}

		err := v.validateParam(param, table)
		if err != nil {
			return fmt.Errorf(`invalid parameter "%s": %w`, param.Name, err)
		}
	}
	for _, where := range where {
		if _, ok := inputNames[where.Name]; ok {
			return violation(errorCode900, fmt.Errorf(`duplicate parameter/where-clause name "%s"`, where.Name))
		}
		inputNames[where.Name] = struct{}{}

		err := v.validateWhere(where, table)
		if err != nil {
			return fmt.Errorf(`invalid where-clause "%s": %w`, where.Name, err)
		}
	}

	return nil
}

/*
###################################################################################################

	Input: 1000-1099

###################################################################################################
*/

// both validateParam and validateWhere use the validateInput function, but validateWhere needs
// additional checks for the operator

func (v *Validator) validateParam(p *databases.Parameter[*spec.KwilAny], table *databases.Table[*spec.KwilAny]) error {
	return v.validateInput(p, table)
}

func (v *Validator) validateWhere(where *databases.WhereClause[*spec.KwilAny], table *databases.Table[*spec.KwilAny]) error {
	if !where.Operator.IsValid() {
		return violation(errorCode1008, fmt.Errorf(`unknown operator: %d`, where.Operator.Int()))
	}

	col := table.GetColumn(where.Column)
	if col == nil {
		return violation(errorCode1001, fmt.Errorf(`column does not exist "%s"`, where.Column))
	}

	if !operatorCanBeOnColumnType(where.Operator, col.Type) {
		return violation(errorCode1009, fmt.Errorf(`operator "%s" can not be used on column type "%s"`, where.Operator.String(), table.GetColumn(where.Column).Type.String()))
	}

	return v.validateInput(where, table)
}

func (v *Validator) validateInput(input databases.Input[*spec.KwilAny], table *databases.Table[*spec.KwilAny]) error {
	if err := CheckName(input.GetName(), MAX_INPUT_NAME_LENGTH); err != nil {
		return violation(errorCode1000, fmt.Errorf(`invalid input name: %w`, err))
	}

	col := table.GetColumn(input.GetColumn())
	if col == nil {
		return violation(errorCode1001, fmt.Errorf(`column does not exist "%s"`, input.GetColumn()))
	}

	// check that modifier is valid
	if !input.GetModifier().IsValid() {
		return violation(errorCode1010, fmt.Errorf(`unknown modifier: %d`, input.GetModifier().Int()))
	}

	if input.GetStatic() {
		// check if value type matches column type
		if col.Type != input.GetValue().Type() && !input.GetValue().IsEmpty() {
			return violation(errorCode1002, fmt.Errorf(`value "%s" must be of type "%s" for parameter on column "%s"`, fmt.Sprint(input.GetValue()), col.Type.String(), input.GetColumn()))
		}

		if err := v.validateCallerModifier(input, col); err != nil {
			return fmt.Errorf(`invalid caller modifier: %w`, err)
		}

	} else { // not static: users can fill in the value
		if input.GetValue() != nil { // double nested to avoid nil pointer dereference
			if !input.GetValue().IsEmpty() {
				return violation(errorCode1002, fmt.Errorf(`value must not be set for non-static parameter / where-clause on column "%s"`, input.GetColumn()))
			}
		}

		if input.GetModifier() == spec.CALLER {
			return violation(errorCode1004, fmt.Errorf(`modifier CALLER can not be on non-static parameter / where-clause "%s"`, input.GetColumn()))
		}
	}

	return nil
}

// provides validations if the modifier is caller
func (v *Validator) validateCallerModifier(input databases.Input[*spec.KwilAny], col *databases.Column[*spec.KwilAny]) error {
	if input.GetModifier() != spec.CALLER {
		return nil
	}

	if !input.GetValue().IsEmpty() {
		return violation(errorCode1003, fmt.Errorf(`value must not be set for caller modifier on column "%s". received: %s`, input.GetColumn(), input.GetValue().String()))
	}

	if !input.GetStatic() {
		return violation(errorCode1004, fmt.Errorf(`parameter must be static for caller modifier on column "%s"`, input.GetColumn()))
	}

	if col.Type != spec.STRING {
		return violation(errorCode1005, fmt.Errorf(`column type must be string for caller modifier on column "%s"`, input.GetColumn()))
	}

	min := col.GetAttribute(spec.MIN_LENGTH)
	if min != nil {
		minVal, err := min.Value.AsInt()
		if err != nil {
			// this shiuld already have been caught by the attribute validation
			return fmt.Errorf("unexpected error.  could not convert min length to int while validating queries: %w", err)
		}

		if minVal > MIN_WALLET_LENGTH {
			return violation(errorCode1006, fmt.Errorf(`column "%s" min length is greater than than shorted supported address length %d`, input.GetColumn(), MIN_WALLET_LENGTH))
		}
	}

	max := col.GetAttribute(spec.MAX_LENGTH)
	if max != nil {
		maxVal, err := max.Value.AsInt()
		if err != nil {
			// this shiuld already have been caught by the attribute validation
			return fmt.Errorf("unexpected error.  could not convert max length to int while validating queries: %w", err)
		}

		if maxVal < MAX_WALLET_LENGTH {
			return violation(errorCode1007, fmt.Errorf(`column "%s" max length is less than than longest supported address length %d`, input.GetColumn(), MAX_WALLET_LENGTH))
		}
	}

	return nil
}

func operatorCanBeOnColumnType(operator spec.ComparisonOperatorType, colType spec.DataType) bool {
	if colType.IsNumeric() {
		// currently all operators work here but this might change in the future
		return operator == spec.EQUAL || operator == spec.NOT_EQUAL || operator == spec.GREATER_THAN || operator == spec.GREATER_THAN_OR_EQUAL || operator == spec.LESS_THAN || operator == spec.LESS_THAN_OR_EQUAL
	}

	if colType.IsText() {
		return operator == spec.EQUAL || operator == spec.NOT_EQUAL
	}

	if colType == spec.BOOLEAN {
		return operator == spec.EQUAL || operator == spec.NOT_EQUAL
	}

	if colType == spec.UUID {
		return operator == spec.EQUAL || operator == spec.NOT_EQUAL
	}

	return false
}
