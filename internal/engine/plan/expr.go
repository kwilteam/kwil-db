package plan

type LogicalExprResolver interface {
	ResolveExpr(LogicalPlan) *field
}

type LogicalExpr interface {
	ToField(input LogicalPlan) field
	String() string
}

// The Column expression simply represents a reference to a named column
type Column struct {
	Name string
}

func (c *Column) ToField(input LogicalPlan) *field {
	for _, f := range input.Schema().fields {
		if f.ColName == c.Name {
			return f
		}
	}

	return nil
}

func (c *Column) String() string {
	return c.Name
}
