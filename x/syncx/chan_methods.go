package syncx

import (
	"context"
	"kwil/x"
)

func Map[T, R any](source Chan[T], dest Chan[R], fn func(T) R) {
	for v := range source.Read() {
		if !dest.Write(fn(v)) {
			break
		}
	}
}

func CopyTo[T any](source Chan[T], dest Chan[T]) {
	for v := range source.Read() {
		if !dest.Write(v) {
			break
		}
	}
}

func FanOut[T any](ctx context.Context, source Chan[T], destinations ...Chan[T]) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v, ok := <-source.Read():
			if !ok {
				return nil
			}
			cnt := len(destinations)
			for _, dest := range destinations {
				if !dest.Write(v) {
					cnt--
				}
			}
			if cnt == 0 {
				return nil
			}
		}
	}
}

func MapAsync[T, R any](source Chan[T], dest Chan[R], fn func(T) R) <-chan x.Void {
	ch := make(chan x.Void)

	go func(s Chan[T], d Chan[R], f func(T) R, c chan x.Void) {
		Map(s, d, f)
		close(c)
	}(source, dest, fn, ch)

	return ch
}

func CopyToAsync[T any](source Chan[T], dest Chan[T]) <-chan x.Void {
	ch := make(chan x.Void)

	go func(s Chan[T], d Chan[T], c chan x.Void) {
		CopyTo(s, d)
		close(c)
	}(source, dest, ch)

	return ch
}

func FanOutAsync[T any](ctx context.Context, source Chan[T], destinations ...Chan[T]) <-chan error {
	ch := make(chan error)

	go func(s Chan[T], d []Chan[T], c chan error) {
		err := FanOut(ctx, s, d...)
		if err != nil {
			ch <- err
		}
		close(c)
	}(source, destinations, ch)

	return ch
}
