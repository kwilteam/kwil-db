package client

import (
	"context"
	"fmt"
	pb "kwil/api/protobuf/info/v0/gen/go"
	"kwil/x/types/accounts"
)

// @yaiba TODO: move to other folder
type NodeInfo struct {
	ValidatorAccount string
	FundingPool      string
}

func (c *Client) GetAccount(ctx context.Context, address string) (accounts.Account, error) {
	res, err := c.infoClt.GetAccount(ctx, &pb.GetAccountRequest{
		Address: address,
	})
	if err != nil {
		return accounts.Account{}, fmt.Errorf("failed to get info: %w", err)
	}

	return accounts.Account{
		Address: res.Account.Address,
		Nonce:   res.Account.Nonce,
		Balance: res.Account.Balance,
		Spent:   res.Account.Spent,
	}, nil
}

func (c *Client) GetInfo(ctx context.Context) (NodeInfo, error) {
	res, err := c.infoClt.GetInfo(ctx, &pb.GetInfoRequest{})
	if err != nil {
		return NodeInfo{}, fmt.Errorf("failed to get info: %w", err)
	}
	return NodeInfo{
		ValidatorAccount: res.ValidatorAccount,
		FundingPool:      res.FundingPool,
	}, nil
}
