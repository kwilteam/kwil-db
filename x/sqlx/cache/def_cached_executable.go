package cache

import (
	"fmt"
	spec "kwil/x/sqlx"
	"kwil/x/sqlx/models"
)

type Executable struct {
	Name      string
	Statement string
	Table     string
	Type      spec.QueryType

	Args       []*Arg
	UserInputs []*UserInput
}

type Arg struct {
	Position      int
	Static        bool
	InputPosition int
	Value         any
	Type          spec.DataType
	Modifier      spec.ModifierType
}

func (a *Arg) From(m *models.Arg) error {
	a.Position = m.Position
	a.Static = m.Static
	a.InputPosition = m.InputPosition
	a.Value = m.Value
	a.Type = m.Type
	a.Modifier = m.Modifier
	return nil
}

type UserInput struct {
	Position int
	Type     spec.DataType
	Value    any
}

func (u *UserInput) From(m *models.UserInput) error {
	u.Position = m.Position
	u.Type = m.Type
	u.Value = m.Value
	return nil
}

func (c *Executable) From(m *models.ExecutableQuery) error {
	typ, err := spec.Conversion.ConvertQueryType(m.Type)
	if err != nil {
		return fmt.Errorf("failed to convert query type: %s", err.Error())
	}

	for _, arg := range m.Args {
		a := &Arg{}
		if err := a.From(arg); err != nil {
			return err
		}
		c.Args = append(c.Args, a)
	}

	for _, usrInpt := range m.UserInputs {
		u := &UserInput{}
		if err := u.From(usrInpt); err != nil {
			return err
		}
		c.UserInputs = append(c.UserInputs, u)
	}

	c.Name = m.Name
	c.Statement = m.Statement
	c.Table = m.Table
	c.Type = typ
	return nil
}

// PrepareInputs takes user inputs, and converts them into a slice of any that can be executed against the database
func (q *Executable) PrepareInputs(sender string, usrInpts []*models.UserInput) ([]any, error) {
	// convert the user inputs to a map for easier lookup
	usrInptsMap := make(map[int]*models.UserInput)
	for _, usrInpt := range usrInpts {
		usrInptsMap[usrInpt.Position] = usrInpt
	}

	// loop through args and fill in their values
	returns := make([]any, len(q.Args))
	for _, arg := range q.Args {

		// if the arg is static, just set the value
		if arg.Static {
			defVal, err := arg.determineDefault(sender)
			if err != nil {
				return nil, fmt.Errorf(`invalid default for arg "%d": %w`, arg.Position, err)
			}

			returns[arg.Position] = defVal
			continue
		}

		// if not static, the arg must contain a corresponding user input
		usrInpt, ok := usrInptsMap[arg.InputPosition]
		if !ok {
			return nil, fmt.Errorf(`missing user input for arg "%d"`, arg.Position)
		}

		// check that the user input type matches the arg type
		if usrInpt.Type != arg.Type {
			return nil, fmt.Errorf(`invalid user input for arg "%d": expected type "%s", got "%s"`, arg.Position, arg.Type.String(), usrInpt.Type.String())
		}

		// convert the user input value to the arg type
		converted, err := spec.Conversion.StringToAnyGolangType(usrInpt.Value, arg.Type)
		if err != nil {
			return nil, fmt.Errorf(`failed to convert user input for arg "%d": %w`, arg.Position, err)
		}

		returns[arg.Position] = converted
	}

	return returns, nil
}

// setDefault will return the default value for the arg.
// If the arg has a modifier, it will apply that accordingly
func (a *Arg) determineDefault(sender string) (any, error) {
	if a.Modifier == spec.NO_MODIFIER {
		return a.Value, nil
	}
	if a.Modifier == spec.CALLER {
		return sender, nil
	}

	return nil, fmt.Errorf(`invalid modifier "%s" for default value`, a.Modifier.String())
}
