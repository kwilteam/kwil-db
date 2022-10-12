package ksl

import (
	"fmt"
	"strings"
)

func Format(node any) string {
	switch node := node.(type) {
	case *Arg:
		return Format(node.Value)
	case *Kwarg:
		return fmt.Sprintf("%s=%s", Format(node.Name), Format(node.Value))
	case *BlockAnnotation:
		var t string
		for _, trivia := range node.LeadingTrivia.Trivia {
			t += Format(trivia)
		}
		return t + Format(node.Annotation)
	case *Annotation:
		args := make([]string, len(node.Args))
		for i, arg := range node.Args {
			args[i] = Format(arg)
		}
		kwargs := make([]string, len(node.Kwargs))
		for i, kwarg := range node.Kwargs {
			kwargs[i] = Format(kwarg)
		}
		if len(args) == 0 && len(kwargs) == 0 {
			return fmt.Sprintf("@%s", Format(node.Name))
		}
		return fmt.Sprintf("@%s(%s)", Format(node.Name), strings.Join(append(args, kwargs...), ", "))

	case *Slice:
		values := make([]string, len(node.Values))
		for i, value := range node.Values {
			values[i] = Format(value)
		}
		return fmt.Sprintf("[%s]", strings.Join(values, ", "))

	case *FunctionCall:
		args := make([]string, len(node.Args))
		for i, arg := range node.Args {
			args[i] = Format(arg)
		}
		kwargs := make([]string, len(node.Kwargs))
		for i, kwarg := range node.Kwargs {
			kwargs[i] = Format(kwarg)
		}
		return fmt.Sprintf("%s(%s)", Format(node.Name), strings.Join(append(args, kwargs...), ", "))
	case *Bool:
		if node.Value {
			return "true"
		}
		return "false"
	case *Var:
		return fmt.Sprintf("$%s", Format(node.Name))
	case *Name:
		parts := make([]string, 0, len(node.Qualifiers)+1)
		for _, qualifier := range node.Qualifiers {
			parts = append(parts, Format(qualifier))
		}
		parts = append(parts, Format(node.Value))
		return strings.Join(parts, ".")
	case *Str:
		return fmt.Sprintf(`"%s"`, node.Value)
	case *Expr:
		return fmt.Sprintf("`%s`", node.Value)
	case *Heredoc:
		return fmt.Sprintf("<<%s\n\t%s\n%s", node.Token, node.Value, node.Token)
	case *Directive:
		var t string
		for _, trivia := range node.LeadingTrivia.Trivia {
			t += Format(trivia)
		}
		parts := make([]string, 0, 8)
		parts = append(parts, "@"+Format(node.Kind))
		if node.Name != nil {
			parts = append(parts, fmt.Sprintf("%s =", Format(node.Name)))
		}
		parts = append(parts, Format(node.Value))
		return t + strings.Join(parts, " ")

	case *Ident:
		return node.Value
	case *Float:
		return fmt.Sprintf("%f", node.Value)
	case *Int:
		return fmt.Sprintf("%d", node.Value)
	case *Type:
		var val string
		if node.IsArray {
			val = "[]"
		}
		val += Format(node.Name)
		if node.Size != nil {
			val += fmt.Sprintf("(%s)", Format(node.Size))
		}
		if node.Nullable {
			val += "?"
		}
		return val

	case *KeyValue:
		if node.Value == nil {
			return Format(node.Key)
		}
		return fmt.Sprintf("%s = %s", Format(node.Key), Format(node.Value))
	case *Property:
		var t string
		for _, trivia := range node.LeadingTrivia.Trivia {
			t += Format(trivia)
		}
		vals := make([]string, 0, 8)
		if node.Key != nil {
			vals = append(vals, Format(node.Key))
		}
		if node.Value != nil {
			vals = append(vals, Format(node.Value))
		}
		return t + strings.Join(vals, " = ")
	case *Declaration:
		val := Format(node.Name) + ": " + Format(node.Type)
		attrs := make([]string, len(node.Annotations))
		for i, attr := range node.Annotations {
			attrs[i] = Format(attr)
		}
		attrstr := strings.Join(attrs, " ")
		return strings.Join([]string{val, attrstr}, " ")
	case *Modifier:
		var vals []string
		if node.Keyword != nil {
			vals = append(vals, Format(node.Keyword))
		}
		if node.Target != nil {
			vals = append(vals, Format(node.Target))
		}
		return strings.Join(vals, " ")
	case *Labels:
		var vals []string
		for _, label := range node.Values {
			vals = append(vals, Format(label))
		}
		return "[" + strings.Join(vals, ", ") + "]"
	case *Comment:
		return "/// " + node.Value + "\n"
	case *Newline:
		return node.Value
	case *Resource:
		var val string
		for _, trivia := range node.LeadingTrivia.Trivia {
			val += Format(trivia)
		}
		val += Format(node.Kind)
		if node.Name != nil {
			val += " " + Format(node.Name)
		}
		if node.Modifier != nil {
			val += " " + Format(node.Modifier)
		}
		if node.Labels != nil {
			val += " " + Format(node.Labels)
		}
		val += " {"
		for _, child := range node.Fields {
			val += strings.ReplaceAll("\n"+Format(child), "\n", "\n\t")
		}
		val += "\n}"
		return val
	case *Document:
		vals := make([]string, len(node.Entries))
		for i, val := range node.Entries {
			vals[i] = Format(val)
		}
		return strings.Join(vals, "\n")
	default:
		panic(fmt.Sprintf("unknown node type %T", node))
	}
}
