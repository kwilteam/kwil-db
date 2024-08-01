package virtual_plan

import (
	"context"
	"fmt"
	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

const (
	SeqScanRowCost   = 100
	IndexScanRowCost = 5
	ProjectionCost   = 10
	FilterEqCost     = 20
)

type VSeqScanOp struct {
	ds         ds.DataSource
	projection []string
}

func (s *VSeqScanOp) String() string {
	return fmt.Sprintf("VSeqScan: schema=%s, projection=%s",
		s.ds.Schema(), s.projection)
}

func (s *VSeqScanOp) Schema() *datatypes.Schema {
	return s.ds.Schema().Project(s.projection...)
}

func (s *VSeqScanOp) Inputs() []VirtualPlan {
	return []VirtualPlan{}
}

func (s *VSeqScanOp) Execute(ctx context.Context) *ds.Result {
	return s.ds.Scan(ctx, s.projection...)
}

func (s *VSeqScanOp) Statistics() *datatypes.Statistics {
	return &datatypes.Statistics{}
}

func (s *VSeqScanOp) Cost() int64 {
	return SeqScanRowCost * s.Statistics().RowCount
}

func VSeqScan(datasource ds.DataSource, projection ...string) VirtualPlan {
	return &VSeqScanOp{ds: datasource, projection: projection}
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
	return IndexScanRowCost
}

func VIndexScan(dataSrc ds.DataSource, projection ...string) VirtualPlan {
	return &VIndexScanOp{ds: dataSrc, projection: projection}
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
	return p.input.Cost() + ProjectionCost
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
	return s.input.Cost() + FilterEqCost
}

func VSelection(input VirtualPlan, expr VirtualExpr) VirtualPlan {
	return &VFilterOp{input: input, expr: expr}
}

// VSortOp represents a sort operation.
// NOTE: this is only a stub implementation.
type VSortOp struct {
	input VirtualPlan
	exprs []VirtualExpr
}

func (s *VSortOp) String() string {
	exprsStr := make([]string, 0, len(s.exprs))
	for _, expr := range s.exprs {
		exprsStr = append(exprsStr, expr.Resolve(s.input))
	}
	return fmt.Sprintf("VSort: %s", exprsStr)
}

func (s *VSortOp) Schema() *datatypes.Schema {
	return s.input.Schema()
}

func (s *VSortOp) Inputs() []VirtualPlan {
	return []VirtualPlan{s.input}
}

func (s *VSortOp) Execute(ctx context.Context) *ds.Result {
	input := s.input.Execute(ctx)
	return ds.ResultFromStream(s.input.Schema(), input.Stream)
}

func (s *VSortOp) Statistics() *datatypes.Statistics {
	return s.input.Statistics()
}

func (s *VSortOp) Cost() int64 {
	return s.input.Cost()
}

func VSortSTUB(input VirtualPlan, expr ...VirtualExpr) VirtualPlan {
	return &VSortOp{input: input, exprs: expr}
}
