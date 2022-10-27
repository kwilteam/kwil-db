package wallet

import (
	"kwil/x"
	"kwil/x/async"
	"kwil/x/cfgx"
	"kwil/x/svcx/messaging/pub"
	"kwil/x/svcx/messaging/sub"
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

func NewRequestProcessor(cfg cfgx.Config) (RequestProcessor, error) {
	p, err := pub.NewByteEmitterSingleClient(cfg.Select("wallet-confirmation-publisher"))
	if err != nil {
		return nil, err
	}

	c, err := sub.NewTransientReceiver(cfg.Select("wallet-request-consumer"))
	if err != nil {
		return nil, err
	}

	return &request_processor{p, c, make(chan x.Void), make(chan x.Void), &sync.WaitGroup{}, &sync.Mutex{}, false}, nil
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
