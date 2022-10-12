package messaging

//import (
//	"context"
//	"fmt"
//	"kwil/x/messaging/pub"
//)
//
//func FromCtx[T any](ctx context.Context, emitter_id string) (pub.Emitter[T], error) {
//	if emitter_id == "" {
//		return nil, fmt.Errorf("emitter id cannot be empty")
//	}
//
//	e, ok := ctx.Value(emitter_id).(Emitter[T])
//	if !ok {
//		return nil, pub.ErrEmitterNotFound()
//	}
//
//	return e, nil
//}
//
//func StoreCtx[T any](ctx context.Context, emitter pub.Emitter[T]) (context.Context, error) {
//	if emitter == nil {
//		return nil, fmt.Errorf("emitter is nil")
//	}
//
//	if emitter.ID() == "" {
//		return nil, fmt.Errorf("emitter id cannot be empty")
//	}
//
//	if ctx == nil {
//		ctx = context.Background()
//	}
//
//	return context.WithValue(ctx, emitter.ID(), emitter), nil
//}
//
//func StoreCtxWithAlias[T any](ctx context.Context, emitter Emitter[T], alias string) (context.Context, error) {
//	if alias == "" {
//		return nil, fmt.Errorf("alias cannot be empty")
//	}
//
//	if emitter == nil {
//		return nil, fmt.Errorf("emitter is nil")
//	}
//
//	if ctx == nil {
//		ctx = context.Background()
//	}
//
//	return context.WithValue(ctx, alias, emitter), nil
//}
