package service

import "context"

func (s *depositsService) Start(ctx context.Context) error {
	s.log.Info("starting deposits service")
	s.mu.Lock()
	defer s.mu.Unlock()

	// start the chain synchronizer
	return s.chain.Start(ctx)
}
