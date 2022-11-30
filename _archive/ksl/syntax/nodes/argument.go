package nodes

import "ksl"

type ArgumentList struct {
	Arguments Arguments
	Span      ksl.Range
}

func (a ArgumentList) Range() ksl.Range { return a.Span }
func (a ArgumentList) Args() Arguments  { return a.Arguments }

func (a ArgumentList) ArgsKwargs() ([]*Argument, map[string]*Argument) { return a.Arguments.All() }

func (a *ArgumentList) Arg(name string) (*Argument, bool) {
	if a == nil {
		return nil, false
	}
	return a.Arguments.Get(name)
}

type Arguments []*Argument

func (a Arguments) Get(name string) (*Argument, bool) {
	if a == nil {
		return nil, false
	}

	for _, arg := range a {
		if arg.GetName() == name {
			return arg, true
		}
	}
	return nil, false
}

func (a Arguments) All() ([]*Argument, map[string]*Argument) {
	if a == nil {
		return nil, nil
	}

	var args []*Argument
	m := make(map[string]*Argument, len(a))
	for _, arg := range a {
		if arg.GetName() == "" {
			args = append(args, arg)
		} else {
			m[arg.GetName()] = arg
		}
	}
	return args, m
}

type Argument struct {
	Name  *Name
	Value Expression
	Span  ksl.Range
}

func (a Argument) Range() ksl.Range    { return a.Span }
func (a *Argument) GetName() string    { return a.Name.String() }
func (a *Argument) GetNameNode() *Name { return a.Name }
func (a *Argument) Identifier() *Name  { return a.Name }
