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

	r := &request_Service{p, c, sync.Mutex{}, make(map[string]async.Action)}

	c.OnSpent(r.onSpent)
	c.OnWithdrawn(r.onWithdrawn)

	return r, nil
}
