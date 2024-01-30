package plan

import (
	"bytes"
	"fmt"
)

type Statistic struct{}

type CostEstimate interface {
	Estimate(Statistic) float64
}

// field provides the name and data type for a field within a schema.
type field struct {
	OriginalTblName string
	OriginalColName string
	TblName         string
	ColName         string
	DB              string

	Type string
	//Nullable bool

	used bool
}

func (f *field) String() string {
	return fmt.Sprintf("%s.%s", f.TblName, f.ColName)
}

// index is a list of fields that are used to index a schema.
type index []*field

// schema is the metadata for a data source or the results from a query.
// A schema consists of one or more fields.
type schema struct {
	tblName  string
	tblAlias string
	fields   []*field
	idxs     []index
}

func (s *schema) Select(projection []string) *schema {
	var fields []*field
	for _, name := range projection {
		for _, f := range s.fields {
			if f.OriginalColName == name {
				fields = append(fields, f)
				break
			}
		}
	}
	return &schema{
		fields: fields,
	}
}

func newSchema(cols ...*field) *schema {
	return &schema{
		fields: cols,
	}
}

type DataSource interface {
	// Schema Return the schema for the underlying data source
	Schema() *schema
	// Scan Scan the data source, select the specified columns
	Scan(projection []string) []any
}

type memDataSource struct {
	schema *schema
}

func (ds *memDataSource) Schema() *schema {
	return ds.schema
}

func (ds *memDataSource) Scan(projection []string) []any {
	return nil
}

type Plan interface {
	Schema() *schema
	//SetSchema(*schema)
}

//type basePlan struct {
//	schema *schema
//}
//
//func (p *basePlan) Schema() *schema {
//	return p.schema
//}
//
//func (p *basePlan) SetSchema(s *schema) {
//	p.schema = s
//}

type LogicalPlan interface {
	Plan
	fmt.Stringer

	Inputs() []LogicalPlan
	Accept(LogicalVisitor) any
}

//
//func format(plan LogicalPlan, indent int) string {
//	var msg bytes.Buffer
//	for i := 0; i < indent; i++ {
//		msg.WriteString(" ")
//	}
//	msg.WriteString(plan.String())
//	msg.WriteString("\n")
//	for _, child := range plan.Inputs() {
//		msg.WriteString(format(child, indent+2))
//	}
//	return msg.String()
//}

func explain(p LogicalPlan) string {
	return explainWithPrefix(p, "", "")
}

func explainWithPrefix(p LogicalPlan, titlePrefix string, bodyPrefix string) string {
	var msg bytes.Buffer
	msg.WriteString(titlePrefix)

	ov := NewExplainVisitor()
	msg.WriteString(p.Accept(ov).(string))
	msg.WriteString("\n")

	for _, child := range p.Inputs() {
		msg.WriteString(explainWithPrefix(
			child,
			bodyPrefix+"->  ",
			bodyPrefix+"      "))
	}
	return msg.String()

}

type LogicalVisitor interface {
	Visit(LogicalPlan) any
	VisitLogicalScan(*LogicalScan) any
	VisitLogicalProjection(*LogicalProjection) any
	VisitLogicalFilter(*LogicalFilter) any
	VisitLogicalJoin(*LogicalJoin) any
	VisitLogicalLimit(*LogicalLimit) any
	VisitLogicalAggregate(*LogicalAggregate) any
	VisitLogicalSort(*LogicalSort) any
	VisitLogicalDistinct(*LogicalDistinct) any
	VisitLogicalSet(*LogicalSet) any
}
