package ty

import "github.com/samber/mo"

func MapOption[T, U any](o mo.Option[T], f func(T) mo.Option[U]) mo.Option[U] {
	if o.IsAbsent() {
		return mo.None[U]()
	}
	return f(o.MustGet())
}
