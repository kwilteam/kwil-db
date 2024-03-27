package datasource

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type ColumnValue interface {
	Type() string
	Value() any
}

type LiteralColumnValue struct {
	value any
}

func (c *LiteralColumnValue) Type() string {
	return fmt.Sprintf("%T", c.value)
}

func (c *LiteralColumnValue) Value() any {
	return c.value
}

func NewLiteralColumnValue(v any) *LiteralColumnValue {
	return &LiteralColumnValue{value: v}
}

type Row []ColumnValue

func (r Row) String() string {
	var cols []string
	for _, c := range r {
		cols = append(cols, fmt.Sprintf("%v", c.Value()))
	}
	return fmt.Sprintf("[%s]", strings.Join(cols, ", "))
}

type RowPipeline chan Row

func newRowPipeline(rows []Row) RowPipeline {
	out := make(RowPipeline)
	go func() {
		defer close(out)

		for _, r := range rows {
			out <- r
		}
	}()
	return out

}

type Result struct {
	schema *datatypes.Schema
	stream RowPipeline
}

func ResultFromStream(s *datatypes.Schema, rows RowPipeline) *Result {
	return &Result{schema: s, stream: rows}
}

func ResultFromRaw(s *datatypes.Schema, rows []Row) *Result {
	// TODO: use RowPipeline all the way
	return &Result{schema: s, stream: newRowPipeline(rows)}
}

func (r *Result) Schema() *datatypes.Schema {
	return r.schema
}

func (r *Result) Next() (Row, bool) {
	row, ok := <-r.stream
	return row, ok
}

func (r *Result) ToCsv() string {
	var sb strings.Builder
	for _, f := range r.schema.Fields {
		sb.WriteString(fmt.Sprintf("%s", f.Name))
		if f != r.schema.Fields[len(r.schema.Fields)-1] {
			sb.WriteString(",")
		}
	}

	sb.WriteString("\n")

	for {
		row, ok := <-r.stream
		if !ok {
			break
		}
		for i, col := range row {
			sb.WriteString(fmt.Sprintf("%v", col.Value()))
			if i < len(row)-1 {
				sb.WriteString(",")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// tap converts an array to a channel(stream of data)
func tap[T any](ctx context.Context, in []T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, element := range in {
			select {
			case <-ctx.Done():
				return
			case out <- element:
			}
		}
	}()
	return out
}

// filter applies a function to each element in the input channel
func filter[T any](ctx context.Context, in <-chan T, fn func(T) bool) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for i := range in {
			if !fn(i) {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case out <- i:
			}
		}
	}()
	return out
}

// smap applies a function to each element in the input channel
func smap[T any](ctx context.Context, in <-chan T, fn func(T) T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for i := range in {
			select {
			case <-ctx.Done():
				return
			case out <- fn(i):
			}
		}
	}()
	return out
}

// transform applies a function to each element in the input channel
func transform[I any, O any](ctx context.Context, in <-chan I, fn func(I) O) <-chan O {
	out := make(chan O)
	go func() {
		defer close(out)
		for i := range in {
			select {
			case <-ctx.Done():
				return
			case out <- fn(i):
			}
		}
	}()
	return out
}

// collect collects all the elements in the input channel
func collect[T any](ctx context.Context, in <-chan T) []T {
	out := make([]T, 0)
	for element := range in {
		select {
		case <-ctx.Done():
			return out
		default:
			out = append(out, element)
		}
	}
	return out
}

// SteamingAPI defines a streaming API.
type SteamingAPI[T any] interface {
	Transform(ctx context.Context, fn func(T) T) SteamingAPI[T]
	Filter(ctx context.Context, fn func(T) bool) SteamingAPI[T]
	Collect(ctx context.Context) []T
}

// SAI is a StreamingApi Implementation.
type SAI[T any] struct {
	dataIn <-chan T
}

func Tap[T any](ctx context.Context, in []T) SteamingAPI[T] {
	return &SAI[T]{dataIn: tap(ctx, in)}
}

func (s SAI[T]) Transform(ctx context.Context, fn func(T) T) SteamingAPI[T] {
	return &SAI[T]{dataIn: transform(ctx, s.dataIn, fn)}
}

func (s SAI[T]) Filter(ctx context.Context, fn func(T) bool) SteamingAPI[T] {
	return &SAI[T]{dataIn: filter(ctx, s.dataIn, fn)}
}

func (s SAI[T]) Collect(ctx context.Context) []T {
	var out []T
	for {
		select {
		case <-ctx.Done():
			return out
		case v, ok := <-s.dataIn:
			if !ok {
				return out
			}
			out = append(out, v)
		}
	}
}

//// SteamingAPI defines a streaming API.
//type SteamingAPI[I any, O any] interface {
//	//Transform(ctx context.Context, fn func(I) O) SteamingAPI[I, O]
//	//Map(ctx context.Context, fn func(I) I) SteamingAPI[I, O]
//	//Filter(ctx context.Context, fn func(I) bool) SteamingAPI[I, O]
//	Collect(ctx context.Context) []I
//}
//
//// SAI is a StreamingApi Implementation.
//type SAI[I any, O any] struct {
//	dataIn <-chan I
//}
//
//func Tap[I any, O any](ctx context.Context, in []I) SteamingAPI[I, O] {
//	return &SAI[I, O]{dataIn: tap(ctx, in)}
//}
//
////func (s *SAI[I any, O any]) Transform(ctx context.Context, fn func(I) O) SteamingAPI[I, O] {
////	return &SAI[I,O]{dataIn: transform(ctx, s.dataIn, fn)}
////}
////
////
////func (s *SAI[I any, O any]) Map(ctx context.Context, fn func(I) I) SteamingAPI[I, O] {
////	return &SAI[I,O]{dataIn: smap(ctx, s.dataIn, fn)}
////}
//
////func (s *SAI[I any, O any]) Filter(ctx context.Context, fn func(I) bool) SteamingAPI[I, O] {
////	return &SAI[I, O]{dataIn: filter(ctx, s.dataIn, fn)}
////}
//
//func (s SAI[I any, O any]) Collect(ctx context.Context) []I {
//	var out []I
//	for {
//		select {
//		case <-ctx.Done():
//			return out
//		case v, ok := <-s.dataIn:
//			if !ok {
//				return out
//			}
//			out = append(out, v)
//		}
//	}
//}
