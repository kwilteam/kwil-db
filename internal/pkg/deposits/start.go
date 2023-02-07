package deposits

import "context"

func (s *depositer) Start(ctx context.Context) error {
	s.log.Sugar().Info("starting deposits service")

	// start the chain synchronizer
	return s.chain.Start(ctx)
}
