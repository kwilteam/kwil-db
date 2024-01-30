package query

import (
	"bytes"
)

type field struct {
	Name     string
	Type     string
	Nullable bool
}

type schema struct {
	Fields []*field
	Keys   [][]*field
}

func (s *schema) Select(projection []string) *schema {
	var fields []*field
	for _, name := range projection {
		for _, f := range s.Fields {
			if f.Name == name {
				fields = append(fields, f)
				break
			}
		}
	}
	return &schema{
		Fields: fields,
	}
}

type DataSource interface {
	// Schema Return the schema for the underlying data source
	Schema() *schema
	// Scan Scan the data source, select the specified columns
	Scan(projection []string) []any
}

type LogicalPlan interface {
	Schema() *schema
	Children() []LogicalPlan
	String() string
}

func format(plan LogicalPlan, indent int) string {
	var msg bytes.Buffer
	for i := 0; i < indent; i++ {
		msg.WriteString(" ")
	}
	msg.WriteString(plan.String())
	msg.WriteString("\n")
	for _, child := range plan.Children() {
		msg.WriteString(format(child, indent+2))
	}
	return msg.String()
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
	for _, f := range input.Schema().Fields {
		if f.Name == c.Name {
			return f
		}
	}

	return nil
}

func (c *Column) String() string {
	return c.Name
}

type ScanPlan struct {
	Source     DataSource
	Projection []string

	schema *schema
}

func NewScanPlan(source DataSource, projection []string) *ScanPlan {
	p := &ScanPlan{
		Source:     source,
		Projection: projection,
	}

	p.schema = p.deriveSchema()
	return p
}

func (p *ScanPlan) deriveSchema() *schema {
	return p.Source.Schema().Select(p.Projection)
}

func (p *ScanPlan) String() string {
	if p.Projection == nil {
		return "Scan: projection=*"
	} else {
		return "Scan: projection=" + p.Projection.String()
	}
}

func (p *ScanPlan) Schema() *schema {
	return p.schema
}

func (p *ScanPlan) Children() []LogicalPlan {
	return nil
}

type ProjectionPlan struct {
	input LogicalPlan
	Exprs []LogicalExpr
}
