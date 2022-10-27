package wallet

import (
	"kwil/x/async"
	"kwil/x/cfgx"
	"kwil/x/svcx/messaging/pub"
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

	c, err := newRequestEvents(cfg.Select("wallet-request-processor"))
	if err != nil {
		return nil, err
	}

	return &request_processor{p, c}, nil
}
