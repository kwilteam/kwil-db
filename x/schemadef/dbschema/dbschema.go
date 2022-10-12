package dbschema

type (
	Role struct {
		Name    string
		Default bool
		Queries []*Query
	}

	Query struct {
		Name      string
		Statement string
		Inputs    []*Input
		Output    *Output
	}

	Input struct {
		Name     string
		Position int
		Type     string
	}

	Output struct {
		Name     string
		Type     string
		Nullable bool
	}
)
