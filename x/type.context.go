package x

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

var empty_map = make(map[string]interface{}) // should be used for *internal consumption only*
var errContextIdEmpty = fmt.Errorf("lookup id cannot be empty")
var requestIdentityFn RequestInjectorFn = func(r *http.Request) *http.Request {
	return r
}
var contextIdentityFn ContextInjectorFn = func(ctx context.Context) context.Context {
	return ctx
}

// ErrContextIdEmpty is returned when a context id is empty.
func ErrContextIdEmpty() error { return errContextIdEmpty }

// ContextInjectorFn is a function that can be used to inject
// data into a context.
type ContextInjectorFn func(ctx context.Context) context.Context

// RequestInjectorFn is a function that can be used to inject
// data into a request.
type RequestInjectorFn func(r *http.Request) *http.Request

// InjectItem is a single item that has been injected into
// a request.
type InjectItem struct {
	ID   string
	Item interface{}
	next *InjectItem
}

// Injectable is a convenience function for creating InjectItem
func Injectable[T any](id string, item T) InjectItem {
	if id == "" {
		panic(errContextIdEmpty)
	}

	return InjectItem{ID: id, Item: item}
}

// Append will append the given items to the option set.
func (i InjectItem) Append(id string, item interface{}) *InjectItem {
	return &InjectItem{id, item, &i}
}

// ContextInjector returns a function that can be used to inject
// data into a context, and a function that can be used to remove
// the injected data.
func ContextInjector(options ...*InjectItem) (ContextInjectorFn, Clearable[*InjectItem]) {
	if len(options) == 0 {
		return contextIdentityFn, &item_Iterator{m: empty_map, pos: -1}
	}

	m := make(map[string]interface{}, len(options))
	for _, opt := range options {
		m[opt.ID] = opt.Item
	}

	mm := &m
	mm_copy := &m

	iter := &item_Iterator{m: *mm_copy}

	return func(ctx context.Context) context.Context {
		return &multi_value_context{mm, ctx}
	}, iter
}

// AsRequestInjector is a convenience method for simplifying composition
// of the use of RequestInjector.
func (i InjectItem) AsRequestInjector() (RequestInjectorFn, Clearable[*InjectItem]) {
	next := &i

	var items []*InjectItem
	for next != nil {
		items = append(items, &InjectItem{ID: next.ID, Item: next.Item})
		next = next.next
	}

	return RequestInjector(items...)
}

// AsContextInjector is a convenience method for simplifying composition
// of the use of ContextInjector.
func (i InjectItem) AsContextInjector() (ContextInjectorFn, Clearable[*InjectItem]) {
	next := &i

	var items []*InjectItem
	for next != nil {
		items = append(items, &InjectItem{ID: next.ID, Item: next.Item})
		next = next.next
	}

	return ContextInjector(items...)
}

// RequestInjector returns a function that can be used to inject
// data into a request, and a function that can be used to remove
// the injected data.
func RequestInjector(options ...*InjectItem) (RequestInjectorFn, Clearable[*InjectItem]) {
	if len(options) == 0 {
		return requestIdentityFn, &item_Iterator{m: empty_map, pos: -1}
	}

	m := make(map[string]interface{}, len(options))
	for _, opt := range options {
		m[opt.ID] = opt.Item
	}

	mm := &m
	mm_copy := &m

	iter := &item_Iterator{m: *mm_copy}

	return func(r *http.Request) *http.Request {
		ctx := &multi_value_context{mm, r.Context()}
		return r.WithContext(ctx)
	}, iter
}

// Resolve will unwrap the value from context given id and value. If
// the context is nil, the golang default for type T will be returned.
// If the id is empty, the method will panic.
func Resolve[T any](ctx context.Context, id string) (out T) {
	if id == "" {
		panic(errContextIdEmpty)
	}

	if ctx == nil {
		return
	}

	e, ok := ctx.Value(id).(T)
	if !ok {
		return
	}

	return e
}

// ResolveOrDefault will unwrap the value from context given id and value.
// If the context is nil, the defaultValue param will be returned. If the
// id is empty, the method will panic.
func ResolveOrDefault[T any](ctx context.Context, id string, defaultValue T) T {
	if id == "" {
		panic(errContextIdEmpty)
	}

	e, ok := ctx.Value(id).(T)
	if !ok {
		return defaultValue
	}

	return e
}

// Inject will wrap the context with the given id and value. If the context
// is nil, context.Background() will be used. If the id is empty, the
// method will panic.
func Inject[T any](ctx context.Context, id string, item T) context.Context {
	if id == "" {
		panic(errContextIdEmpty)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, id, item)
}

type multi_value_context struct {
	m   *map[string]interface{}
	ctx context.Context
}

func (m *multi_value_context) Deadline() (deadline time.Time, ok bool) {
	return m.ctx.Deadline()
}

func (m *multi_value_context) Done() <-chan struct{} {
	return m.ctx.Done()
}

func (m *multi_value_context) Err() error {
	return m.Err()
}

func (m *multi_value_context) Value(key interface{}) interface{} {
	if m.ctx.Err() == nil {
		*m.m = empty_map
		return m.ctx.Value(key)
	}

	mm := *m.m
	if key != nil {
		id, ok := key.(string)
		if ok {
			v, ok := mm[id]
			if !ok {
				return v
			}
		}
	}

	return m.ctx.Value(key)
}

type item_Iterator struct {
	m       map[string]interface{}
	pos     int
	items   []*InjectItem
	current *InjectItem
}

func (t *item_Iterator) HasNext() bool {
	if t.pos == -1 {
		return false
	}

	if t.items == nil {
		var items []*InjectItem
		for k, v := range t.m {
			items = append(items, &InjectItem{ID: k, Item: v})
		}
		t.items = items
	}

	if t.pos >= len(t.items) {
		t.pos = -1
		t.m = empty_map
		return false
	}

	t.pos++
	t.current = t.items[t.pos]

	return true
}

func (t *item_Iterator) Value() *InjectItem {
	return t.current
}

func (t *item_Iterator) Clear() Iterator[*InjectItem] {
	return t
}
