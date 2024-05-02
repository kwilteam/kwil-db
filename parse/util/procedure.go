package util

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/metadata"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

// FindProcOrForeign finds a procedure or foreign procedure by name.
// it returns the parameter types, and the return type.
func FindProcOrForeign(schema *types.Schema, name string) (parameters []*types.DataType, returns *types.ProcedureReturn, err error) {
	if proc, ok := schema.FindProcedure(name); ok {
		for _, p := range proc.Parameters {
			parameters = append(parameters, p.Type)
		}

		if proc.Returns != nil {
			returns = proc.Returns
		}

		return parameters, returns, nil
	}

	if proc, ok := schema.FindForeignProcedure(name); ok {
		parameters = append(parameters, proc.Parameters...)

		if proc.Returns != nil {
			returns = proc.Returns
		}

		return parameters, returns, nil
	}

	return nil, nil, fmt.Errorf("%w: %s", parseTypes.ErrUnknownFunctionOrProcedure, name)
}

// FormatProcedureName formats a procedure name for usage in postgres. This
// simply prepends the name with _fp_
func FormatForeignProcedureName(name string) string {
	return fmt.Sprintf("_fp_%s", name)
}

// FormatParameterName formats a parameter name for usage in postgres. This
// simply prepends the name with _param_, and removes the $ prefix.
func FormatParameterName(name string) string {
	return fmt.Sprintf("_param_%s", name[1:])
}

// UnformatParameterName removes the _param_ prefix from a parameter name.
// If it does not have the prefix, it will return the name as is.
func UnformatParameterName(name string) string {
	if len(name) > 7 && name[:7] == "_param_" {
		return "$" + name[7:]
	}

	return name
}

// FormatContextualVariableName formats a contextual variable name for usage in postgres.
// This uses the current_setting function to get the value of the variable. It also
// removes the @ prefix. If the type is not a text type, it will also type cast it.
// The type casting is necessary since current_setting returns all values as text.
func FormatContextualVariableName(name string, dataType *types.DataType) string {
	str := fmt.Sprintf("current_setting('%s.%s')", metadata.PgSessionPrefix, name[1:])
	if dataType.Equals(types.TextType) {
		return str
	}

	switch dataType {
	case types.BlobType:
		return fmt.Sprintf("%s::bytea", str)
	case types.IntType:
		return fmt.Sprintf("%s::int8", str)
	case types.BoolType:
		return fmt.Sprintf("%s::bool", str)
	case types.UUIDType:
		return fmt.Sprintf("%s::uuid", str)
	}

	panic("unallowed contextual variable type: " + dataType.String())
}

// UnformatContextualVariableName removes the current_setting function from a contextual variable name.
// If it does not have the function, it will return the name as is. It will also remove any type casting.
func UnformatContextualVariableName(name string) string {
	// trim the type casting
	strs := strings.Split(name, "::")
	name = strs[0]

	if len(name) > 17 && name[:17] == "current_setting('" {
		return "@" + name[17:len(name)-2]
	}

	return name
}
