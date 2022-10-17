package pub

import (
	"fmt"
	"kwil/x/async"
)

var errEmitterNotFound = fmt.Errorf("emitter not found")

func ErrEmitterNotFound() error {
	return errEmitterNotFound
}

type AckNackFn func(err error) async.Action

var none_ack AckNackFn = func(err error) async.Action {
	return none_action
}

var none_action = async.CompletedAction()

func ackAsync() (AckNackFn, async.Action) {
	action := async.NewAction()
	return func(err error) async.Action {
		action.CompleteOrFail(err)
		return none_action
	}, action
}
