package host

import (
	"context"
	"time"

	types "kwil/x/common/host/service"
)

type serviceContextImpl struct {
	types.ServiceContext
	ctx      context.Context
	fnById   func(id int32) (types.Service, error)
	fnByName func(name string) (types.Service, error)
}

func newServiceContext(ctx context.Context, fnById func(id int32) (types.Service, error), fnByName func(name string) (types.Service, error)) types.ServiceContext {
	return &serviceContextImpl{ctx: ctx, fnById: fnById, fnByName: fnByName}
}

func (s *serviceContextImpl) GetServiceById(id int32) (types.Service, error) {
	return s.fnById(id)
}

func (s *serviceContextImpl) GetServiceByName(name string) (types.Service, error) {
	return s.fnByName(name)
}

func (s *serviceContextImpl) Value(key interface{}) interface{} {
	return s.ctx.Value(key)
}

// Done() <-chan struct{}
func (s *serviceContextImpl) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s *serviceContextImpl) Err() error {
	return s.ctx.Err()
}

func (s *serviceContextImpl) Deadline() (deadline time.Time, ok bool) {
	return s.ctx.Deadline()
}
