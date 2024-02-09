package virtual_plan

import (
	"fmt"

	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
)

type VScanOp struct {
	ds         ds.DataSource
	projection []string
}

func (s *VScanOp) String() string {
	return fmt.Sprintf("VScan: schema=%s, projection=%s",
		s.ds.Schema(), s.projection)
}

func (s *VScanOp) Schema() *ds.Schema {
	return s.ds.Schema().Select(s.projection...)
}

func (s *VScanOp) Inputs() []VirtualPlan {
	return []VirtualPlan{}
}

func (s *VScanOp) Execute() *ds.Result {
	return s.ds.Scan(s.projection...)
}

func VScan(ds ds.DataSource, projection ...string) VirtualPlan {
	return &VScanOp{ds: ds, projection: projection}
}

type VProjectionOp struct {
	input  VirtualPlan
	exprs  []VirtualExpr
	schema *ds.Schema
}

func (p *VProjectionOp) String() string {
	exprsStr := make([]string, 0, len(p.exprs))
	for _, expr := range p.exprs {
		exprsStr = append(exprsStr, expr.Resolve(p.input))
	}
	return fmt.Sprintf("VProjection: %s", exprsStr)

	//return fmt.Sprintf("VProjection: %s", p.exprs)
}

func (p *VProjectionOp) Schema() *ds.Schema {
	return p.schema
}

func (p *VProjectionOp) Inputs() []VirtualPlan {
	return []VirtualPlan{p.input}
}

func (p *VProjectionOp) Execute() *ds.Result {
	input := p.input.Execute()
	out := make(ds.RowPipeline)

	go func() {
		defer close(out)

		for {
			row, ok := input.Next()
			if !ok {
				break
			}

			cols := make(ds.Row, 0, len(row))
			for _, expr := range p.exprs {
				cols = append(cols, expr.evaluate(row))
			}

			out <- cols
		}
	}()
	return ds.ResultFromStream(p.schema, out)
}

func VProjection(input VirtualPlan, schema *ds.Schema, exprs ...VirtualExpr) VirtualPlan {
	return &VProjectionOp{input: input, exprs: exprs, schema: schema}
}

type VSelectionOp struct {
	input VirtualPlan
	expr  VirtualExpr
}

func (s *VSelectionOp) String() string {
	return fmt.Sprintf("VSelection: %s", s.expr.Resolve(s.input))
	//return fmt.Sprintf("VSelection: %s", s.expr)
}

func (s *VSelectionOp) Schema() *ds.Schema {
	return s.input.Schema()
}

func (s *VSelectionOp) Inputs() []VirtualPlan {
	return []VirtualPlan{s.input}
}

func (s *VSelectionOp) Execute() *ds.Result {
	input := s.input.Execute()

	out := make(ds.RowPipeline)

	go func() {
		defer close(out)

		for {
			row, ok := input.Next()
			if !ok {
				break
			}

			if s.expr.evaluate(row).Value().(bool) {
				out <- row
			}
		}
	}()

	return ds.ResultFromStream(s.input.Schema(), out)
}

func VSelection(input VirtualPlan, expr VirtualExpr) VirtualPlan {
	return &VSelectionOp{input: input, expr: expr}
}
