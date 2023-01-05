package service

import "context"

func (s *depositsService) Sync(ctx context.Context) error {
	s.log.Info("starting deposits service")

	// start the chain synchronizer
	return s.chain.Start(ctx)
}
