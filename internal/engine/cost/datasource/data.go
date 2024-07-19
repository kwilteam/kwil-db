package datasource

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

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
	Schema *datatypes.Schema
	Stream <-chan Row
}

func ResultFromStream(s *datatypes.Schema, stream <-chan Row) *Result {
	return &Result{Schema: s, Stream: stream}
}

func ResultFromRaw(s *datatypes.Schema, rows []Row) *Result {
	// TODO: use RowPipeline all the way
	return &Result{Schema: s, Stream: newRowPipeline(rows)}
}

func (r *Result) ToCsv() string {
	var sb strings.Builder
	for _, f := range r.Schema.Fields {
		sb.WriteString(f.Name)
		if f != r.Schema.Fields[len(r.Schema.Fields)-1] {
			sb.WriteString(",")
		}
	}

	sb.WriteString("\n")

	for {
		row, ok := <-r.Stream
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

// StreamTap converts an array to a channel(stream of data)
func StreamTap[T any](ctx context.Context, in []T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, element := range in {
			select {
			case <-ctx.Done():
				return
			case out <- element:
				//fmt.Printf("tapped: %v\n", element)
			}
		}
	}()
	return out
}

// StreamFilter applies a function to each element in the input channel
func StreamFilter[T any](ctx context.Context, in <-chan T, fn func(T) bool) <-chan T {
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
				//fmt.Printf("filtered: %v\n", i)
			}
		}
	}()
	return out
}

// StreamMap applies a function to each element in the input channel
func StreamMap[T any](ctx context.Context, in <-chan T, fn func(T) T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for i := range in {
			select {
			case <-ctx.Done():
				return
			case out <- fn(i):
				//fmt.Printf("mapped: %v\n", i)
			}
		}
	}()
	return out
}

// StreamTransform applies a function to each element in the input channel.
// It transforms the element to another type
func StreamTransform[I any, O any](ctx context.Context, in <-chan I, fn func(I) O) <-chan O {
	out := make(chan O)
	go func() {
		defer close(out)
		for i := range in {
			select {
			case <-ctx.Done():
				return
			case out <- fn(i):
				//fmt.Printf("transformed: %v\n", i)
			}
		}
	}()
	return out
}

// StreamCollect collects all the elements in the input channel
func StreamCollect[T any](ctx context.Context, in <-chan T) []T {
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
// NOTE: it can only handle one type of data, thus no Transform method.
type SteamingAPI[T any] interface {
	Map(ctx context.Context, fn func(T) T) SteamingAPI[T]
	Filter(ctx context.Context, fn func(T) bool) SteamingAPI[T]
	Collect(ctx context.Context) []T
}

// SAI is a StreamingApi Implementation.
type SAI[T any] struct {
	dataIn <-chan T
}

// ToStream converts an array to a channel(stream of data)
func ToStream[T any](ctx context.Context, in []T) SteamingAPI[T] {
	return &SAI[T]{dataIn: StreamTap(ctx, in)}
}

func (s SAI[T]) Map(ctx context.Context, fn func(T) T) SteamingAPI[T] {
	return &SAI[T]{dataIn: StreamMap(ctx, s.dataIn, fn)}
}

func (s SAI[T]) Filter(ctx context.Context, fn func(T) bool) SteamingAPI[T] {
	return &SAI[T]{dataIn: StreamFilter(ctx, s.dataIn, fn)}
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
