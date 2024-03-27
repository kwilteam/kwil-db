package virtual_plan

import (
	"context"
	"fmt"

	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type VTableScanOp struct {
	ds         ds.DataSource
	projection []string
}

func (s *VTableScanOp) String() string {
	return fmt.Sprintf("VTableScan: schema=%s, projection=%s",
		s.ds.Schema(), s.projection)
}

func (s *VTableScanOp) Schema() *datatypes.Schema {
	return s.ds.Schema().Project(s.projection...)
}

func (s *VTableScanOp) Inputs() []VirtualPlan {
	return []VirtualPlan{}
}

func (s *VTableScanOp) Execute(ctx context.Context) *ds.Result {
	return s.ds.Scan(ctx, s.projection...)
}

func (s *VTableScanOp) Statistics() *datatypes.Statistics {
	return &datatypes.Statistics{}
}

func (s *VTableScanOp) Cost() int64 {
	return 0
}

func VTableScan(datasource ds.SchemaSource, projection ...string) VirtualPlan {
	ds := ds.SchemaSourceToDataSource(datasource)
	return &VTableScanOp{ds: ds, projection: projection}
}

type VIndexScanOp struct {
	ds         ds.DataSource
	projection []string
}

func (s *VIndexScanOp) String() string {
	return fmt.Sprintf("VIndexScan: schema=%s, projection=%s",
		s.ds.Schema(), s.projection)
}

func (s *VIndexScanOp) Schema() *datatypes.Schema {
	return s.ds.Schema().Project(s.projection...)
}

func (s *VIndexScanOp) Inputs() []VirtualPlan {
	return []VirtualPlan{}
}

func (s *VIndexScanOp) Execute(ctx context.Context) *ds.Result {
	return s.ds.Scan(ctx, s.projection...)
}

func (s *VIndexScanOp) Statistics() *datatypes.Statistics {
	return &datatypes.Statistics{}
}

func (s *VIndexScanOp) Cost() int64 {
	return 0
}

func VIndexScan(datasource ds.SchemaSource, projection ...string) VirtualPlan {
	ds := ds.SchemaSourceToDataSource(datasource)
	return &VIndexScanOp{ds: ds, projection: projection}
}

type VProjectionOp struct {
	input  VirtualPlan
	exprs  []VirtualExpr
	schema *datatypes.Schema
}

func (p *VProjectionOp) String() string {
	exprsStr := make([]string, 0, len(p.exprs))
	for _, expr := range p.exprs {
		exprsStr = append(exprsStr, expr.Resolve(p.input))
	}
	return fmt.Sprintf("VProjection: %s", exprsStr)

	//return fmt.Sprintf("VProjection: %s", p.exprs)
}

func (p *VProjectionOp) Schema() *datatypes.Schema {
	return p.schema
}

func (p *VProjectionOp) Inputs() []VirtualPlan {
	return []VirtualPlan{p.input}
}

func (p *VProjectionOp) Execute(ctx context.Context) *ds.Result {
	input := p.input.Execute(ctx)

	out := ds.StreamMap(ctx, input.Stream, func(row ds.Row) ds.Row {
		cols := make(ds.Row, 0, len(row))
		for _, expr := range p.exprs {
			cols = append(cols, expr.evaluate(row))
		}
		return cols
	})

	return ds.ResultFromStream(p.schema, out)
}

func (p *VProjectionOp) Statistics() *datatypes.Statistics {
	return p.input.Statistics()
}

func (p *VProjectionOp) Cost() int64 {
	return p.input.Cost()
}

func VProjection(input VirtualPlan, schema *datatypes.Schema, exprs ...VirtualExpr) VirtualPlan {
	return &VProjectionOp{input: input, exprs: exprs, schema: schema}
}

type VFilterOp struct {
	input VirtualPlan
	expr  VirtualExpr
}

func (s *VFilterOp) String() string {
	return fmt.Sprintf("VFilter: %s", s.expr.Resolve(s.input))
	//return fmt.Sprintf("VSelection: %s", s.expr)
}

func (s *VFilterOp) Schema() *datatypes.Schema {
	return s.input.Schema()
}

func (s *VFilterOp) Inputs() []VirtualPlan {
	return []VirtualPlan{s.input}
}

func (s *VFilterOp) Execute(ctx context.Context) *ds.Result {
	input := s.input.Execute(ctx)

	out := ds.StreamFilter(ctx, input.Stream, func(row ds.Row) bool {
		res := s.expr.evaluate(row).Value()
		return res.(bool)
	})

	return ds.ResultFromStream(s.input.Schema(), out)
}

func (s *VFilterOp) Statistics() *datatypes.Statistics {
	return s.input.Statistics()
}

func (s *VFilterOp) Cost() int64 {
	return s.input.Cost()
}

func VSelection(input VirtualPlan, expr VirtualExpr) VirtualPlan {
	return &VFilterOp{input: input, expr: expr}
}
