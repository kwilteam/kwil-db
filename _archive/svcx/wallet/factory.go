package wallet

import (
	"kwil/_archive/svcx/messaging/mx"
	"kwil/_archive/svcx/messaging/pub"
	"kwil/_archive/svcx/messaging/sub"
	"kwil/x"
	"kwil/x/async"
	"kwil/x/cfgx"
	"sync"
)

func NewRequestService(cfg cfgx.Config) (RequestService, error) {
	p, err := pub.NewByteEmitterSingleClient(cfg.Select("wallet-request-publisher"))
	if err != nil {
		return nil, err
	}

	c, err := newConfirmationEvents(cfg.Select("wallet-confirmation-consumer"))
	if err != nil {
		return nil, err
	}

	r := &request_Service{p, c, sync.Mutex{}, make(map[string]async.Action)}
	c.handler = r.handle_event_response

	return r, nil
}

func NewRequestProcessor(cfg cfgx.Config, transform MessageTransform) (RequestProcessor, error) {
	p, err := pub.NewByteEmitterSingleClient(cfg.Select("wallet-confirmation-publisher"))
	if err != nil {
		return nil, err
	}

	c, err := sub.NewTransientReceiver(cfg.Select("wallet-request-consumer"))
	if err != nil {
		return nil, err
	}

	if transform == nil {
		transform = SyncTransform(func(msg *mx.RawMessage) (*mx.RawMessage, error) {
			return msg, nil
		})
	}

	return &request_processor{
		p:         p,
		e:         c,
		done:      make(chan x.Void),
		stop:      make(chan x.Void),
		transform: transform,
		wg:        &sync.WaitGroup{},
		mu:        &sync.Mutex{}}, nil
}

func newConfirmationEvents(cfg cfgx.Config) (*confirmation_events, error) {
	e, err := sub.NewTransientReceiver(cfg)
	if err != nil {
		return nil, err
	}

	return &confirmation_events{
		e:    e,
		wg:   sync.WaitGroup{},
		stop: make(chan x.Void),
		done: make(chan x.Void),
		mu:   sync.Mutex{},
	}, nil
}

func SyncTransform(fn func(*mx.RawMessage) (*mx.RawMessage, error)) MessageTransform {
	return func(msg *mx.RawMessage) async.Task[*mx.RawMessage] {
		msg, err := fn(msg)
		if err != nil {
			return async.FailedTask[*mx.RawMessage](err)
		}

		return async.CompletedTask(msg)
	}
}
