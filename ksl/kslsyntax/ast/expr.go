package ast

import (
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"ksl"
)

var _ Expr = (*Number)(nil)
var _ Expr = (*Float)(nil)
var _ Expr = (*Int)(nil)
var _ Expr = (*Str)(nil)
var _ Expr = (*QuotedStr)(nil)
var _ Expr = (*Heredoc)(nil)
var _ Expr = (*Bool)(nil)
var _ Expr = (*Null)(nil)
var _ Expr = (*Object)(nil)
var _ Expr = (*List)(nil)
var _ Expr = (*Var)(nil)
var _ Expr = (*FunctionCall)(nil)

type Expr interface {
	Node
	ksl.Expression
}

func (n *Number) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	return cty.NumberVal(n.Value), nil
}

func (n *Float) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	return cty.NumberFloatVal(n.GetFloat()), nil
}

func (n *Int) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	return cty.NumberIntVal(n.GetInt64()), nil
}

func (n *Str) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	return cty.StringVal(n.GetString()), nil
}

func (n *QuotedStr) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	return cty.StringVal(n.GetString()), nil
}

func (n *Heredoc) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	lines := make([]string, 0, len(n.Values))
	for _, line := range n.Values {
		if n.StripIndent {
			lines = append(lines, strings.TrimLeft(line.GetString(), " \t"))
		} else {
			lines = append(lines, line.GetString())
		}
	}
	return cty.StringVal(strings.Join(lines, "")), nil
}

func (n *Bool) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	return cty.BoolVal(n.GetBool()), nil
}

func (n *Null) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	return cty.NullVal(cty.DynamicPseudoType), nil
}

func (n *Object) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	attrs := map[string]cty.Value{}
	for _, kv := range n.GetAttributes() {
		val, valDiags := kv.GetValue().Eval(ctx)
		diags = append(diags, valDiags...)
		attrs[kv.Name.Value] = val
	}

	return cty.ObjectVal(attrs), diags
}

func (n *List) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	vals := make([]cty.Value, 0, len(n.Values))
	for _, v := range n.GetValues() {
		val, valDiags := v.Eval(ctx)
		diags = append(diags, valDiags...)
		vals = append(vals, val)
	}
	return cty.ListVal(vals), diags
}

func (n *Var) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	v, ok := ctx.Variables[n.Name]
	if !ok {
		return cty.NilVal, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  "Undefined variable",
			Detail:   fmt.Sprintf("The variable %q is not defined.", n.Name),
			Subject:  n.Range().Ptr(),
		}}
	}

	return v, nil
}

func (n *FunctionCall) Eval(ctx *ksl.Context) (cty.Value, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	fn, ok := ctx.Functions[n.GetName()]
	if !ok {
		return cty.NilVal, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  "Undefined function",
			Detail:   fmt.Sprintf("The function %q is not defined.", n.GetName()),
			Subject:  n.Range().Ptr(),
		}}
	}

	var args []cty.Value
	if n.ArgList != nil {
		args = make([]cty.Value, 0, len(n.ArgList.Args)+len(n.ArgList.Kwargs))
		for _, arg := range n.ArgList.Args {
			val, valDiags := arg.Eval(ctx)
			diags = append(diags, valDiags...)
			args = append(args, val)
		}

		kwargs := make(map[string]cty.Value, len(n.ArgList.Kwargs))
		for _, kwarg := range n.ArgList.Kwargs {
			val, valDiags := kwarg.GetValue().Eval(ctx)
			diags = append(diags, valDiags...)
			kwargs[kwarg.Name.Value] = val
		}

		if len(kwargs) > 0 {
			args = append(args, cty.ObjectVal(kwargs))
		}
	}
	result, err := fn.Call(args)
	if err != nil {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Error calling function",
			Detail:   err.Error(),
			Subject:  n.Range().Ptr(),
		})
	}
	return result, diags
}
