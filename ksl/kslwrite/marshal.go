package kslwrite

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

func Marshal(w io.Writer, f *File) error {
	return formatFile(w, f)
}

func groupDirectives(directives []*Directive) ([]*Directive, []*Directive) {
	var keyed, unkeyed []*Directive

	for _, node := range directives {
		switch {
		case node.Key == "":
			unkeyed = append(unkeyed, node)
		default:
			keyed = append(keyed, node)
		}
	}
	return unkeyed, keyed
}

type byName struct {
	annotations []*Annotation
}

func (b byName) Len() int {
	return len(b.annotations)
}

func (b byName) Less(i, j int) bool {
	return b.annotations[i].Name < b.annotations[j].Name
}

func (b byName) Swap(i, j int) {
	b.annotations[i], b.annotations[j] = b.annotations[j], b.annotations[i]
}

func formatFile(w io.Writer, f *File) error {
	unkeyed, keyed := groupDirectives(f.Directives)

	for _, dir := range unkeyed {
		formatDirective(w, dir)
		fmt.Fprintln(w)
	}
	if len(unkeyed) > 0 {
		fmt.Fprintln(w)
	}
	for _, dir := range keyed {
		formatDirective(w, dir)
		fmt.Fprintln(w)
	}
	if len(keyed) > 0 {
		fmt.Fprintln(w)
	}

	for _, block := range f.Blocks {
		formatBlock(w, block)
		fmt.Fprintln(w)
	}
	return nil
}

func formatAttribute(w io.Writer, a *Attribute) error {
	fmt.Fprintf(w, a.Key)
	if a.Key != "" && a.Value != "" {
		fmt.Fprintf(w, " = ")
	}
	_, err := fmt.Fprintf(w, a.Value)
	return err
}

func formatKwarg(w io.Writer, a *Kwarg) error {
	fmt.Fprintf(w, a.Key)
	if a.Key != "" && a.Value != "" {
		fmt.Fprintf(w, "=")
	}
	_, err := fmt.Fprintf(w, a.Value)
	return err
}

func formatDirective(w io.Writer, d *Directive) error {
	fmt.Fprintf(w, "@%s", d.Name)
	if d.Key != "" || d.Value != "" {
		fmt.Fprintf(w, " ")
	}

	if d.Key != "" {
		fmt.Fprintf(w, "%s", d.Key)
	}
	if d.Key != "" && d.Value != "" {
		fmt.Fprintf(w, " = ")
	}
	if d.Value != "" {
		fmt.Fprintf(w, "%s", d.Value)
	}
	return nil
}

func formatBlock(w io.Writer, b *Block) error {
	fmt.Fprint(w, b.Type)
	if b.Name != "" {
		fmt.Fprintf(w, " %s", b.Name)
	}
	if b.Modifier != "" {
		fmt.Fprintf(w, " %s", b.Modifier)
	}
	if b.Target != "" {
		fmt.Fprintf(w, " %s", b.Target)
	}
	if len(b.Labels) > 0 {
		fmt.Fprint(w, " [")
		for i, l := range b.Labels {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			formatKwarg(w, l)
		}
		fmt.Fprint(w, "]")
	}
	fmt.Fprintf(w, " {")

	if b.Body == nil || len(b.Body.Attributes)+len(b.Body.Annotations)+len(b.Body.Definitions)+len(b.Body.Blocks)+len(b.Body.EnumValues) == 0 {
		_, err := fmt.Fprint(w, "}")
		return err
	} else {
		fmt.Fprintln(w)
		formatBlockBody(newIndentWriter(w), b.Body)
		fmt.Fprint(w, "}")
	}
	return nil
}

func formatBlockBody(w io.Writer, b *BlockBody) error {
	if len(b.Definitions) > 0 {
		wr := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)
		for _, d := range b.Definitions {
			formatDefinition(wr, d)
			fmt.Fprintln(wr)
		}
		wr.Flush()
	}

	if len(b.Attributes) > 0 {
		wr := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)
		for _, a := range b.Attributes {
			formatAttribute(wr, a)
			fmt.Fprintln(wr)
		}
	}

	for _, v := range b.EnumValues {
		fmt.Fprintln(w, v)
	}

	for _, b := range b.Blocks {
		formatBlock(w, b)
		fmt.Fprintln(w)
	}

	sort.Sort(byName{b.Annotations})
	for _, a := range b.Annotations {
		fmt.Fprint(w, "@")
		formatAnnotation(w, a)
		fmt.Fprintln(w)
	}

	return nil
}

func formatAnnotation(w io.Writer, a *Annotation) error {
	fmt.Fprintf(w, "@%s", a.Name)
	if a.Args != nil {
		formatArgList(w, a.Args)
	}
	return nil
}

func formatArgList(w io.Writer, a *ArgList) error {
	fmt.Fprint(w, "(")
	for i, arg := range a.Args {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprint(w, arg)
	}
	if len(a.Args) > 0 && len(a.Kwargs) > 0 {
		fmt.Fprint(w, ", ")
	}
	for i, kwarg := range a.Kwargs {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		formatKwarg(w, kwarg)
	}
	fmt.Fprint(w, ")")
	return nil
}

func formatDefinition(w io.Writer, d *Definition) error {
	fmt.Fprintf(w, "%s:", d.Name)
	fmt.Fprint(w, "\t")
	fmt.Fprintf(w, d.Type)

	if d.IsArray {
		fmt.Fprintf(w, "[]")
	}

	if d.IsOptional {
		fmt.Fprintf(w, "?")
	}

	fmt.Fprint(w, "\t")
	if len(d.Annotations) > 0 {
		sort.Sort(byName{d.Annotations})
		for i, ann := range d.Annotations {
			if i > 0 {
				fmt.Fprint(w, "\n\t\t")
			}
			formatAnnotation(w, ann)
		}
	}
	return nil
}
