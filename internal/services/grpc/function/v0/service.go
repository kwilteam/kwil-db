// package function/v0 implements a grpc server for the Kwil function service.
// the function service is used to remotely execute logic that is determined by
// compile time parameters.
package v0

import (
	"context"

	"github.com/kwilteam/kwil-db/common/ident"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	functionpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/function/v0"
)

type FunctionService struct {
	functionpb.UnimplementedFunctionServiceServer
}

// VerifySignature checks the signature with the given public key and message.
// This only verifies the signature against known kwil-db singing schema, which
// is determined by the signature's type.
func (FunctionService) VerifySignature(_ context.Context,
	req *functionpb.VerifySignatureRequest) (*functionpb.VerifySignatureResponse, error) {
	convSignature := auth.Signature{
		Signature: req.Signature.SignatureBytes,
		Type:      req.Signature.SignatureType,
	}

	err := ident.VerifySignature(req.Sender, req.Msg, &convSignature)
	if err != nil {
		return &functionpb.VerifySignatureResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &functionpb.VerifySignatureResponse{
		Valid: true,
		Error: "",
	}, nil
}
