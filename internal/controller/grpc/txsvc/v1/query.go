package txsvc

import (
	"context"
	"encoding/json"
	"fmt"

	localClient "github.com/cometbft/cometbft/rpc/client/local"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (s *Service) Query(ctx context.Context, req *txpb.QueryRequest) (*txpb.QueryResponse, error) {
	bcClient := localClient.New(s.BcNode)
	req_data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query request data: %w", err)
	}

	res, err := bcClient.ABCIQuery(ctx, "", req_data)
	if err != nil || res.Response.Code != 0 {
		return nil, fmt.Errorf("failed to query with error:  %s and response code: %v", err, res.Response.Code)
	}

	var resp *txpb.QueryResponse
	err = json.Unmarshal(res.Response.Value, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize query response: %w", err)
	}

	return resp, nil
}
