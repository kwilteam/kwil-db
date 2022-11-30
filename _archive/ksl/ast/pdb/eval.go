package pdb

import (
	"fmt"
	"ksl"
	"ksl/syntax/nodes"
	"math/big"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
)

func Eval(expr nodes.Expression, ctx *ksl.Context, val any) ksl.Diagnostics {
	var diags ksl.Diagnostics
	var srcVal cty.Value

	switch v := expr.(type) {
	case *nodes.Number:
		f, _, err := big.ParseFloat(v.Value, 10, 512, big.ToNearestEven)
		if err != nil {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid number",
				Detail:   fmt.Sprintf("Invalid number: %s", v.Value),
				Subject:  v.Range().Ptr(),
			})
		}
		srcVal = cty.NumberVal(f)
	case *nodes.String:
		srcVal = cty.StringVal(v.Value)
	case *nodes.Heredoc:
		lines := make([]string, 0, len(v.Values))
		for _, line := range v.Values {
			if v.StripIndent {
				line = strings.TrimLeft(line, " \t")
			}
			lines = append(lines, line)
		}
		srcVal = cty.StringVal(strings.Join(lines, " "))
	case *nodes.Literal:
		switch v.Value {
		case "true":
			srcVal = cty.True
		case "false":
			srcVal = cty.False
		case "null":
			srcVal = cty.NullVal(cty.DynamicPseudoType)
		default:
			srcVal = cty.StringVal(v.Value)
		}
	case *nodes.Object:
		attrs := map[string]cty.Value{}
		for _, kv := range v.Properties {
			var val cty.Value
			diags = append(diags, Eval(kv.Value, ctx, &val)...)
			attrs[kv.GetName()] = val
		}
		srcVal = cty.ObjectVal(attrs)
	case *nodes.Function:
		fn, ok := ctx.Functions[v.GetName()]
		if !ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Undefined function",
				Detail:   fmt.Sprintf("The function %q is not defined.", v.GetName()),
				Subject:  v.Range().Ptr(),
			})
			return diags
		}
		namedArgs := make(map[string]cty.Value, len(v.Arguments.Arguments))
		for _, arg := range v.GetArgs() {
			var val cty.Value
			diags = append(diags, Eval(arg.Value, ctx, &val)...)
			namedArgs[arg.GetName()] = val
		}
		arg := cty.ObjectVal(namedArgs)
		result, err := fn.Call([]cty.Value{arg})
		if err != nil {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid function call",
				Detail:   err.Error(),
				Subject:  v.Range().Ptr(),
			})
		}
		srcVal = result

	case *nodes.List:
		vals := make([]cty.Value, 0, len(v.Elements))
		for _, v := range v.Elements {
			var value cty.Value
			diags = append(diags, Eval(v, ctx, &value)...)
			vals = append(vals, value)
		}
		srcVal = cty.ListVal(vals)
	case *nodes.Variable:
		val, ok := ctx.Variables[v.GetName()]
		if !ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Undefined variable",
				Detail:   fmt.Sprintf("The variable %q is not defined.", v.GetName()),
				Subject:  v.Range().Ptr(),
			})
		}
		srcVal = val
	}

	convTy, err := gocty.ImpliedType(val)
	if err != nil {
		panic(fmt.Sprintf("unsuitable DecodeExpression target: %s", err))
	}

	if convTy.IsListType() && srcVal.Type().IsPrimitiveType() {
		srcVal = cty.ListVal([]cty.Value{srcVal})
	} else if convTy.IsPrimitiveType() && srcVal.Type().IsListType() {
		listVals := srcVal.AsValueSlice()
		if len(listVals) > 1 {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid type",
				Detail:   "Cannot convert list to primitive type.",
				Subject:  expr.Range().Ptr(),
			})
			return diags
		}
		if len(listVals) == 1 {
			srcVal = listVals[0]
		}
	}

	srcVal, err = convert.Convert(srcVal, convTy)
	if err != nil {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Unsuitable value type",
			Detail:   fmt.Sprintf("Unsuitable value: %s", err.Error()),
			Subject:  expr.Range().Ptr(),
		})
		return diags
	}

	err = gocty.FromCtyValue(srcVal, val)
	if err != nil {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Unsuitable value type",
			Detail:   fmt.Sprintf("Unsuitable value: %s", err.Error()),
			Subject:  expr.Range().Ptr(),
		})
	}

	return diags
}
