package x

import "fmt"

func TryRun(fn Runnable) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if !ok {
				err = fmt.Errorf("unknown panic: %v", r)
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
	}()

	fn()

	return nil
}

func TryCall[T any](fn Callable[T]) (value T, err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if !ok {
				err = fmt.Errorf("unknown panic: %v", r)
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
	}()

	value = fn()

	return
}

func TryAccept[T any](arg T, fn AcceptT[T]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if !ok {
				err = fmt.Errorf("unknown panic: %v", r)
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
	}()

	fn(arg)

	return
}

func TryBiAccept[T, U any](t T, u U, fn BiAccept[T, U]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if !ok {
				err = fmt.Errorf("unknown panic: %v", r)
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
	}()

	fn(t, u)

	return
}

func TryApply[T, R any](arg T, fn ApplyT[T, R]) (value R, err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if !ok {
				err = fmt.Errorf("unknown panic: %v", r)
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
	}()

	value = fn(arg)

	return
}
