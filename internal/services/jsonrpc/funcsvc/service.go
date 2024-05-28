package funcsvc

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	"github.com/kwilteam/kwil-db/core/rpc/json/function"
	"github.com/kwilteam/kwil-db/internal/ident"
	rpcserver "github.com/kwilteam/kwil-db/internal/services/jsonrpc"
	"github.com/kwilteam/kwil-db/internal/version"
)

type Service struct{}

const (
	apiVerMajor = 0
	apiVerMinor = 1
	apiVerPatch = 0
)

var (
	apiSemver = fmt.Sprintf("%d.%d.%d", apiVerMajor, apiVerMinor, apiVerPatch)
)

// The admin Service must be usable as a Svc registered with a JSON-RPC Server.
var _ rpcserver.Svc = (*Service)(nil)

func (svc *Service) Handlers() map[jsonrpc.Method]rpcserver.MethodHandler {
	return map[jsonrpc.Method]rpcserver.MethodHandler{
		function.MethodVersion: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.VersionRequest{}
			return req, func() (any, *jsonrpc.Error) {
				return &jsonrpc.VersionResponse{
					Service:     "function",
					Version:     apiSemver,
					Major:       apiVerMajor,
					Minor:       apiVerMinor,
					Patch:       apiVerPatch,
					KwilVersion: version.KwilVersion,
				}, nil
			}
		},
		function.MethodVerifySig: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &function.VerifySignatureRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.VerifySignature(ctx, req) }
		},
	}
}

// VerifySignature checks the signature with the given public key and message.
// This only verifies the signature against known kwil-db singing schema, which
// is determined by the signature's type.
func (Service) VerifySignature(_ context.Context, req *function.VerifySignatureRequest) (*function.VerifySignatureResponse, *jsonrpc.Error) {
	convSignature := auth.Signature{
		Signature: req.Signature.SignatureBytes,
		Type:      req.Signature.SignatureType,
	}

	err := ident.VerifySignature(req.Sender, req.Msg, &convSignature)
	if err != nil {
		return &function.VerifySignatureResponse{
			Valid:  false,
			Reason: err.Error(),
		}, nil
	}

	return &function.VerifySignatureResponse{
		Valid: true,
	}, nil
}
