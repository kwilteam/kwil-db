package txsvc

import (
	"context"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/ident"
)

// VerifySignature checks the signature with the given public key and message.
// This only verifies the signature against known kwil-db singing schema, which
// is determined by the signature's type.
func (s *Service) VerifySignature(_ context.Context,
	req *txpb.VerifySignatureRequest) (*txpb.VerifySignatureResponse, error) {
	convSignature := auth.Signature{
		Signature: req.Signature.SignatureBytes,
		Type:      req.Signature.SignatureType,
	}

	err := ident.VerifySignature(req.Sender, req.Msg, &convSignature)
	if err != nil {
		return &txpb.VerifySignatureResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &txpb.VerifySignatureResponse{
		Valid: true,
		Error: "",
	}, nil
}
