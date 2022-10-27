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
	p, err := pub.NewByteEmitterSingleClient(cfg.Select("wallet-request-emitter"))
	if err != nil {
		return nil, err
	}

	c, err := newConfirmationEvents(cfg.Select("wallet-confirmation-events"))
	if err != nil {
		return nil, err
	}

	return &request_Service{p, c, sync.Mutex{}, make(map[string]async.Action)}, nil
}

func NewRequestProcessor(cfg cfgx.Config) (RequestProcessor, error) {
	p, err := pub.NewByteEmitterSingleClient(cfg.Select("wallet-confirmation-emitter"))
	if err != nil {
		return nil, err
	}

	c, err := sub.NewTransientReceiver(cfg.Select("wallet-request-processor"))
	if err != nil {
		return nil, err
	}

	return &request_processor{p, c, make(chan x.Void), make(chan x.Void), &sync.WaitGroup{}, &sync.Mutex{}, false}, nil
}
