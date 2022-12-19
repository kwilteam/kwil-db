package client

import "context"

func (e *EVMClient) GetLatestBlock(ctx context.Context) (int64, error) {
	h, err := e.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}

	return h.Number.Int64(), nil
}
