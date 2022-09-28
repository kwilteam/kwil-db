package rx

// Map will execute the function and set the result if the promise
// is successful, else it will propagate the source error
func Map[T, R any](source Listenable[T], fn func(T) R) *Task[R] {
	task := NewTask[R]()

	source.OnComplete(func(v T, err error) {
		if err != nil {
			task.Fail(err)
		} else {
			task.Complete(fn(v))
		}
	})

	return task
}

// FlatMap will execute the function and await the additional promise,
// if the source is errored, the fn will not be called and the error
// will be propagated, else the promise returned by the function will
// be awaited and returned as the value or an error if it results in an
// errored stated
func FlatMap[T, R any](source Listenable[T], fn func(T) Listenable[R]) *Task[R] {
	task := NewTask[R]()

	source.OnComplete(func(v T, e error) {
		if e != nil {
			task.Fail(e)
			return
		}

		fn(v).OnComplete(func(v2 R, e2 error) {
			if e2 != nil {
				task.Complete(v2)
			} else {
				task.Fail(e2)
			}
		})
	})

	return task
}

func Any[T any](sources ...Listenable[T]) *Task[T] {
	task := NewTask[T]()

	for _, source := range sources {
		source.OnComplete(func(v T, e error) {
			task.CompleteOrFail(v, e)
		})
	}

	return task
}

func All[T any](sources ...Listenable[T]) *Task[[]T] {
	if len(sources) == 0 {
		return Success([]T{})
	}

	task := NewTask[[]T]()

	var arr []T
	arrPtr := &arr
	for i := 0; i < len(sources); i++ {
		sources[i].OnComplete(func(v T, err error) {
			if err != nil {
				task.Fail(err)
				return
			}

			larr := append(*arrPtr, v)
			arrPtr = &larr
			if len(larr) == len(sources) {
				task.Complete(arr)
			}
		})
	}

	return task
}
