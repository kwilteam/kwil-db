package interpreter

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/engine/parse"
)

type action struct {
	// Name is the name of the action.
	// It should always be lower case.
	Name string `json:"name"`

	// Parameters are the input parameters of the action.
	Parameters []*namedType `json:"parameters"`
	// Modifiers modify the access to the action.
	Modifiers []precompiles.Modifier `json:"modifiers"`

	// Body is the logic of the action.
	Body []parse.ActionStmt

	// RawStatement is the unparsed CREATE ACTION statement.
	RawStatement string `json:"raw_statement"`

	// Returns specifies the return types of the action.
	Returns *actionReturn `json:"return_types"`
}

func (a *action) GetName() string {
	return a.Name
}

// FromAST sets the fields of the action from an AST node.
func (a *action) FromAST(ast *parse.CreateActionStatement) error {
	a.Name = ast.Name
	a.RawStatement = ast.Raw
	a.Body = ast.Statements

	a.Parameters = convertNamedTypes(ast.Parameters)

	if ast.Returns != nil {
		a.Returns = &actionReturn{
			IsTable: ast.Returns.IsTable,
			Fields:  convertNamedTypes(ast.Returns.Fields),
		}
	}

	modSet := make(map[precompiles.Modifier]struct{})
	a.Modifiers = []precompiles.Modifier{}
	hasPublicPrivateOrSystem := false
	for _, m := range ast.Modifiers {
		mod, err := stringToMod(m)
		if err != nil {
			return err
		}

		if mod == precompiles.PUBLIC || mod == precompiles.PRIVATE || mod == precompiles.SYSTEM {
			if hasPublicPrivateOrSystem {
				return fmt.Errorf("only one of PUBLIC, PRIVATE, or SYSTEM is allowed")
			}

			hasPublicPrivateOrSystem = true
		}

		if _, ok := modSet[mod]; !ok {
			modSet[mod] = struct{}{}
			a.Modifiers = append(a.Modifiers, mod)
		}
	}

	if !hasPublicPrivateOrSystem {
		return fmt.Errorf(`one of PUBLIC, PRIVATE, or SYSTEM access modifier is required. received: "%s"`, strings.Join(ast.Modifiers, ", "))
	}

	return nil
}

// convertNamedTypes converts a list of named types from the AST to the internal representation.
func convertNamedTypes(params []*parse.NamedType) []*namedType {
	namedTypes := make([]*namedType, len(params))
	for i, p := range params {
		namedTypes[i] = &namedType{
			Name: p.Name,
			Type: p.Type,
		}
	}
	return namedTypes
}

// namedType is a parameter in a procedure.
type namedType struct {
	// Name is the name of the parameter.
	// It should always be lower case.
	// If it is a procedure parameter, it should begin
	// with a $.
	Name string `json:"name"`
	// Type is the type of the parameter.
	Type *types.DataType `json:"type"`
}

// actionReturn holds the return type of a procedure.
// EITHER the Type field is set, OR the Table field is set.
type actionReturn struct {
	IsTable bool         `json:"is_table"`
	Fields  []*namedType `json:"fields"`
}

func stringToMod(s string) (precompiles.Modifier, error) {
	switch strings.ToLower(s) {
	case "public":
		return precompiles.PUBLIC, nil
	case "private":
		return precompiles.PRIVATE, nil
	case "system":
		return precompiles.SYSTEM, nil
	case "owner":
		return precompiles.OWNER, nil
	case "view":
		return precompiles.VIEW, nil
	default:
		return "", fmt.Errorf("unknown modifier %s", s)
	}
}
