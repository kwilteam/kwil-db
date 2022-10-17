package async

import "kwil/x"

// Map will execute the function and set the result if the Task
// is successful, else it will propagate the source error.
func Map[T, R any](source Listenable[T], fn func(T) R) Task[R] {
	tk := NewTask[R]()

	source.OnComplete(&Continuation[T]{
		Then: func(value T) {
			tk.Complete(fn(value))
		},
		Catch: func(err error) {
			tk.Fail(err)
		},
	})

	return tk
}

// MapA will execute the function and set the result if the Action
// is successful, else it will propagate the source error.
func MapA[R any](source Action, fn func() R) Task[R] {
	tk := NewTask[R]()

	source.WhenComplete(func(err error) {
		if err != nil {
			tk.Fail(err)
		} else {
			tk.Complete(fn())
		}
	})

	return tk
}

// FlatMap will execute the function and await the additional task.Listenable,
// if the source is errored, the fn will not be called and the error
// will be propagated, else the task.Listenable returned by the function will
// be awaited and returned as the value or an error if it results in an
// errored stated.
func FlatMap[T, R any](source Listenable[T], fn func(T) Listenable[R]) Task[R] {
	tk := NewTask[R]()

	source.OnComplete(fromHandler(func(v T, e error) {
		if e != nil {
			tk.Fail(e)
			return
		}

		fn(v).OnComplete(fromHandler(func(v2 R, e2 error) {
			if e2 != nil {
				tk.Complete(v2)
			} else {
				tk.Fail(e2)
			}
		}))
	}))

	return tk
}

func Any[T any](sources ...Listenable[T]) Task[T] {
	tk := NewTask[T]()

	for _, source := range sources {
		source.OnComplete(fromHandler(func(v T, e error) {
			tk.CompleteOrFail(v, e)
		}))
	}

	return tk
}

func All[T any](sources ...Listenable[T]) Task[[]T] {
	if len(sources) == 0 {
		return CompletedTask([]T{})
	}

	tk := NewTask[[]T]()

	var arr []T
	arrPtr := &arr
	for i := 0; i < len(sources); i++ {
		sources[i].OnComplete(fromHandler(func(v T, err error) {
			if err != nil {
				tk.Fail(err)
				return
			}

			larr := append(*arrPtr, v)
			arrPtr = &larr
			if len(larr) == len(sources) {
				tk.Complete(arr)
			}
		}))
	}

	return tk
}

func fromHandler[T any](fn func(T, error)) *Continuation[T] {
	return &Continuation[T]{
		Then: func(v T) {
			fn(v, nil)
		},
		Catch: func(e error) {
			fn(x.AsDefault[T](), e)
		},
	}
}
