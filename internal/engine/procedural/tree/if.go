package tree

import "strings"

// If is a conditional statement.
type If struct {
	If      *IfThen
	ElseIfs []*IfThen // can be nil
	Else    []Clause  // can be nil
}

func (i *If) MarshalPG(info *SystemInfo) (string, error) {
	str := strings.Builder{}
	str.WriteString("IF ")
	cond, err := i.If.MarshalPG(info)
	if err != nil {
		return "", err
	}
	str.WriteString(cond)

	if len(i.ElseIfs) > 0 {
		for _, e := range i.ElseIfs {
			str.WriteString(" ELSE IF ")
			cond, err := e.MarshalPG(info)
			if err != nil {
				return "", err
			}
			str.WriteString(cond)
		}
	}

	if i.Else != nil {
		str.WriteString(" ELSE\n")
		for _, e := range i.Else {
			cond, err := e.MarshalPG(info)
			if err != nil {
				return "", err
			}
			str.WriteString(cond)
			str.WriteString("\n")

		}
	}

	str.WriteString(" END IF;")

	return str.String(), nil
}

func (i *If) clause() {}

// IfThen is an evaluatable expression that if true, will execute the IfThen clause
type IfThen struct {
	Expr *ExpressionBoolean
	Then []Clause
}

func (b *IfThen) MarshalPG(info *SystemInfo) (string, error) {
	str := strings.Builder{}
	cond, err := b.Expr.MarshalPG(info)
	if err != nil {
		return "", err
	}
	str.WriteString(cond)

	str.WriteString(" THEN\n")
	for _, e := range b.Then {
		cond, err := e.MarshalPG(info)
		if err != nil {
			return "", err
		}
		str.WriteString(cond)
		str.WriteString("\n")
	}

	return str.String(), nil
}
