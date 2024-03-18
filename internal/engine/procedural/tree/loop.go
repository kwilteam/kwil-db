package tree

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/procedural/types"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

type ForLoop struct {
	// Receiver is the variable to receive the looped value.
	// It should be parsed with a $, but that is removed here.
	Receiver *Variable
	// LoopTerm is the thing to loop over.
	LoopTerm loopable
	// Body is the body of the loop.
	Body []Clause
}

func (f *ForLoop) clause() {}

func (f *ForLoop) MarshalPG(info *SystemInfo) (string, error) {
	// plpgsql uses FOR row IN [loopterm] LOOP,
	// except for arrays, which use FOREACH row IN ARRAY [loopterm] LOOP
	str := strings.Builder{}
	_, ok := f.LoopTerm.(*ArrayLoop)
	if ok {
		str.WriteString("FOREACH ")
	} else {
		str.WriteString("FOR ")
	}

	str.WriteString(f.Receiver.Name)
	str.WriteString(" IN ")
	loopterm, err := f.LoopTerm.loopterm(info)
	if err != nil {
		return "", err
	}

	str.WriteString(loopterm)
	str.WriteString(" LOOP\n")

	for _, c := range f.Body {
		cond, err := c.MarshalPG(info)
		if err != nil {
			return "", err
		}
		str.WriteString(cond)
		str.WriteString("\n")
	}

	str.WriteString("\nEND LOOP;")

	return str.String(), nil
}

// loopable defines all types that can be looped over.
type loopable interface {
	// loopterm returns the clause to use in the for loop.
	// this is the FOR row IN [loopterm] LOOP
	loopterm(info *SystemInfo) (string, error)
}

type IntegerRange struct {
	// Start is the start of the range.
	Start int64
	// End is the end of the range.
	End int64
}

func (i *IntegerRange) loopterm(_ *SystemInfo) (string, error) {
	return fmt.Sprintf("%d..%d", i.Start, i.End), nil
}

type SelectLoop struct {
	// Query is the query to loop over.
	Query *tree.SelectStmt
}

func (s *SelectLoop) loopterm(_ *SystemInfo) (str string, err error) {
	// ToSQL can panic, so we need to recover from it.
	defer func() {
		if r := recover(); r != nil {
			err2, ok := r.(error)
			if !ok {
				err2 = fmt.Errorf("%v", r)
			}

			err = err2
		}
	}()

	str = s.Query.ToSQL()
	return str, nil
}

// ArrayLoop is a loop over an array.
type ArrayLoop struct {
	// Array is the array variable to loop over.
	Array *Variable
}

func (a *ArrayLoop) loopterm(info *SystemInfo) (string, error) {
	// check if the variable is an array
	d, ok := info.Context.Variables[a.Array.Name]
	if !ok {
		return "", fmt.Errorf("variable %s not found", a.Array.Name)
	}

	_, ok = d.(*types.ArrayType)
	if !ok {
		return "", fmt.Errorf("variable %s is not an array", a.Array.Name)
	}

	return fmt.Sprintf("ARRAY %s", a.Array.Name), nil
}

type Break struct{}

func (b *Break) clause() {}

func (Break) MarshalPG(_ *SystemInfo) (string, error) {
	return "EXIT;", nil
}
