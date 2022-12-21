package service

import (
	"context"
	"fmt"
	"kwil/x/deposits/chain"
)

func (s *depositsService) WithChainClient(ctx context.Context, chainClient chain.ChainClient) error {
	if s.chainWriter != nil {
		return fmt.Errorf("chain client already set")
	}

	client := chain.New(chainClient, s.db)
	err := client.Sync(ctx)
	if err != nil {
		return err
	}
	err = client.Listen(ctx)
	if err != nil {
		return err
	}

	s.chainWriter = client

	return nil
}
