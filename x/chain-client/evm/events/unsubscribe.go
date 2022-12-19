package events

import "context"

func (e *EVMSubscription) Unsubscribe(ctx context.Context, subscription *EVMSubscription) {
	subscription.sub.Unsubscribe()
}

func (e *EVMSubscription) Err() <-chan error {
	return e.errs
}

func (e *EVMSubscription) Blocks() <-chan int64 {
	return e.blocks
}
