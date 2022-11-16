package spec

import (
	"fmt"
	"ksl"
	"ksl/syntax/ast"

	"golang.org/x/exp/slices"
)

func ValidateDirectives(specs map[string]AnnotSpec, directives ast.Annotations) ksl.Diagnostics {
	visited := map[string]struct{}{}

	var diags ksl.Diagnostics
	for _, dir := range directives {
		dirName := dir.GetName()
		sp, ok := specs[dirName]
		if !ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid directive",
				Detail:   fmt.Sprintf("Directive %q is not a known directive.", dirName),
				Subject:  dir.Range().Ptr(),
			})
			continue
		}
		if sp.Singular {
			if _, ok := visited[dirName]; ok {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Duplicate directive",
					Detail:   fmt.Sprintf("Directive %q is already set.", dirName),
					Subject:  dir.Range().Ptr(),
				})
				continue
			}
			visited[dirName] = struct{}{}
		}

		diags = append(diags, ValidateArgs(sp.Arguments, sp.DefaultArg, dir)...)
	}
	return diags
}

func ValidateArgs(specs map[string]ArgSpec, defaultArg string, container ast.ArgumentHolder) ksl.Diagnostics {
	var diags ksl.Diagnostics

	args := container.GetArgs()
	argMap := make(map[string]*ast.Argument, len(args))
	for _, arg := range args {
		argName := arg.GetName()
		if argName == "" {
			argName = defaultArg
		}

		argSpec, ok := specs[argName]
		if !ok {
			var detail = fmt.Sprintf("Argument %q is not expected here.", argName)
			if argName == "" {
				detail = "Unnamed argument is not expected here."
			}

			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid argument",
				Detail:   detail,
				Subject:  arg.Range().Ptr(),
			})
			continue
		}

		if _, ok := argMap[argName]; ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Duplicate argument",
				Detail:   fmt.Sprintf("Argument %q is already set.", argName),
				Subject:  arg.Range().Ptr(),
			})
			continue
		}
		argMap[argName] = arg
		diags = append(diags, ValidateValue(argSpec.Value, arg.Value)...)
	}
	for _, argSpec := range specs {
		if _, ok := argMap[argSpec.Name]; !ok && argSpec.Required {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Missing argument",
				Detail:   fmt.Sprintf("Argument %q is required.", argSpec.Name),
				Subject:  container.Range().Ptr(),
			})
		}
	}
	return diags
}

func ValidateScalarValue(sp ScalarValueSpec, node ast.Expression) ksl.Diagnostics {
	var diags ksl.Diagnostics
	var kind ScalarKind

	switch node.(type) {
	case *ast.Number:
		kind = NumberKind
	case *ast.String:
		kind = QuotedStringKind
	case *ast.Literal:
		kind = StringLitKind
	case *ast.Heredoc:
		kind = QuotedStringKind
	default:
		kind = NoKind
	}

	if sp.Kind&kind == 0 {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid value",
			Detail:   fmt.Sprintf("Expected a scalar value %s.", GetTypeDescription(sp)),
			Subject:  node.Range().Ptr(),
		})
	}

	return diags
}

func ValidateObject(sp ObjectSpec, e ast.Expression) ksl.Diagnostics {
	var diags ksl.Diagnostics
	if obj, ok := e.(*ast.Object); ok {
		items := make(map[string]ast.Expression, len(obj.Properties))
		for _, kv := range obj.Properties {
			propSpec, ok := sp.Properties[kv.GetName()]
			if !ok {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid property",
					Detail:   fmt.Sprintf("Property %q is not expected here.", kv.GetName()),
					Subject:  kv.Range().Ptr(),
				})
				continue
			}
			if _, ok := items[kv.GetName()]; ok {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Duplicate property",
					Detail:   fmt.Sprintf("Property %q is already set.", kv.GetName()),
					Subject:  kv.Range().Ptr(),
				})
				continue
			}

			diags = append(diags, ValidateValue(propSpec.Value, kv.Value)...)
		}
		for _, propSpec := range sp.Properties {
			if _, ok := items[propSpec.Name]; !ok && propSpec.Required {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Missing property",
					Detail:   fmt.Sprintf("Property %q is required.", propSpec.Name),
					Subject:  e.Range().Ptr(),
				})
			}
		}
	} else {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid value",
			Detail:   "Expected an object.",
			Subject:  e.Range().Ptr(),
		})
	}
	return diags
}

func ValidateList(sp ListSpec, value ast.Expression) ksl.Diagnostics {
	var diags ksl.Diagnostics
	if list, ok := value.(*ast.List); ok {
		for _, elem := range list.Elements {
			diags = append(diags, ValidateValue(sp.ElementType, elem)...)
		}
	} else {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid value",
			Detail:   fmt.Sprintf("Expected a list: %s.", GetTypeDescription(sp)),
			Subject:  value.Range().Ptr(),
		})
	}
	return diags
}

func ValidateOneOf(sp OneOfSpec, value ast.Expression) ksl.Diagnostics {
	for _, opt := range sp.Options {
		if !ValidateValue(opt, value).HasErrors() {
			return nil
		}
	}
	return ksl.Diagnostics{{
		Severity: ksl.DiagError,
		Summary:  "Invalid value",
		Detail:   fmt.Sprintf("Expected one of: %s.", GetTypeDescription(sp)),
		Subject:  value.Range().Ptr(),
	}}
}

func ValidateFunc(sp FuncSpec, value ast.Expression) ksl.Diagnostics {
	var diags ksl.Diagnostics
	if fn, ok := value.(*ast.Function); ok {
		if sp.Name != "" && fn.GetName() != sp.Name {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid value",
				Detail:   fmt.Sprintf("Expected a function like: %s.", GetTypeDescription(sp)),
				Subject:  fn.Range().Ptr(),
			})
			return diags
		}

		args := fn.GetArgs()
		items := make(map[string]*ast.Argument, len(args))
		for _, arg := range args {
			argName := arg.GetName()
			if argName == "" {
				argName = sp.DefaultArg
			}

			argSpec, ok := sp.Arguments[argName]
			if !ok {
				var detail = fmt.Sprintf("Argument %q is not expected here.", argName)
				if argName == "" {
					detail = "Unnamed argument is not expected here."
				}

				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid argument",
					Detail:   detail,
					Subject:  arg.Range().Ptr(),
				})
				continue
			}

			if _, ok := items[argName]; ok {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Duplicate argument",
					Detail:   fmt.Sprintf("Argument %q is already set.", argName),
					Subject:  arg.Range().Ptr(),
				})
				continue
			}
			items[argName] = arg
			diags = append(diags, ValidateValue(argSpec.Value, arg.Value)...)
		}
		for _, argSpec := range sp.Arguments {
			if _, ok := items[argSpec.Name]; !ok && argSpec.Required {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Missing argument",
					Detail:   fmt.Sprintf("Argument %q is required.", argSpec.Name),
					Subject:  value.Range().Ptr(),
				})
			}
		}

	} else {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid value",
			Detail:   fmt.Sprintf("Expected a function like: %s.", GetTypeDescription(sp)),
			Subject:  value.Range().Ptr(),
		})
	}
	return diags
}

func ValidateEnum(sp EnumSpec, value ast.Expression) ksl.Diagnostics {
	var diags ksl.Diagnostics
	var scalar string
	var quoted bool

	switch value := value.(type) {
	case *ast.String:
		scalar = value.Value
		quoted = true
	case *ast.Literal:
		scalar = value.Value
	default:
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid enum value",
			Detail:   fmt.Sprintf("%q is not valid for this context. Expected one of: %s", scalar, GetTypeDescription(sp)),
			Subject:  value.Range().Ptr(),
		})
		return diags
	}

	if !slices.Contains(sp.Values, scalar) {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid enum value",
			Detail:   fmt.Sprintf("%q is not valid for this context. Expected one of: %s", scalar, GetTypeDescription(sp)),
			Subject:  value.Range().Ptr(),
		})
	} else if quoted {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagWarning,
			Summary:  "Unnecessary quotes",
			Detail:   fmt.Sprintf("Quotes around %q are unnecessary.", scalar),
			Subject:  value.Range().Ptr(),
		})
	}
	return diags
}

func ValidateConstantValue(sp ConstantValueSpec, value ast.Expression) ksl.Diagnostics {
	var diags ksl.Diagnostics
	var actual string

	switch value := value.(type) {
	case *ast.Number:
		actual = value.Value
	case *ast.String:
		actual = value.Value
	case *ast.Literal:
		actual = value.Value
	default:
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid value",
			Detail:   fmt.Sprintf("Expected %s.", GetTypeDescription(sp)),
			Subject:  value.Range().Ptr(),
		})
		return diags
	}

	if sp.Value != actual {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Invalid value",
			Detail:   fmt.Sprintf("%q is not a valid value. Expected %s.", actual, GetTypeDescription(sp)),
			Subject:  value.Range().Ptr(),
		})
	}
	return diags
}

func ValidateValue(sp ValueSpec, value ast.Expression) ksl.Diagnostics {
	var diags ksl.Diagnostics
	switch sp := sp.(type) {
	case ScalarValueSpec:
		diags = ValidateScalarValue(sp, value)
	case ObjectSpec:
		diags = ValidateObject(sp, value)
	case ListSpec:
		diags = ValidateList(sp, value)
	case OneOfSpec:
		diags = ValidateOneOf(sp, value)
	case FuncSpec:
		diags = ValidateFunc(sp, value)
	case EnumSpec:
		diags = ValidateEnum(sp, value)
	case ConstantValueSpec:
		diags = ValidateConstantValue(sp, value)
	default:
		panic(fmt.Sprintf("unhandled spec type %T", sp))
	}
	return diags
}
