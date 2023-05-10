package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v0"
)

func (s *Service) GetAddress(ctx context.Context, req *txpb.GetAddressRequest) (*txpb.GetAddressResponse, error) {
	// TODO: once Gavin has fixed config, we should be able to get the node's private key (and thus address)
	// for now, we'll just hardcode it
	return &txpb.GetAddressResponse{
		Address: "0xAfFDC06cF34aFD7D5801A13d48C92AD39609901D",
	}, nil
}
