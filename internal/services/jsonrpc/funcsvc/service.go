package funcsvc

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common/ident"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	"github.com/kwilteam/kwil-db/core/rpc/json/function"
	userjson "github.com/kwilteam/kwil-db/core/rpc/json/user"
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

func verHandler(context.Context, *userjson.VersionRequest) (*userjson.VersionResponse, *jsonrpc.Error) {
	return &userjson.VersionResponse{
		Service:     "function",
		Version:     apiSemver,
		Major:       apiVerMajor,
		Minor:       apiVerMinor,
		Patch:       apiVerPatch,
		KwilVersion: version.KwilVersion,
	}, nil
}

func (svc *Service) Methods() map[jsonrpc.Method]rpcserver.MethodDef {
	return map[jsonrpc.Method]rpcserver.MethodDef{
		function.MethodVersion: rpcserver.MakeMethodDef(verHandler,
			"retrieve the API version of the function service",
			"service info including semver and kwild version"),
		function.MethodVerifySig: rpcserver.MakeMethodDef(svc.VerifySignature,
			"verify a message signature",
			"validity of the signature and any reason for failure"),
	}
}

func (svc *Service) Handlers() map[jsonrpc.Method]rpcserver.MethodHandler {
	handlers := make(map[jsonrpc.Method]rpcserver.MethodHandler)
	for method, def := range svc.Methods() {
		handlers[method] = def.Handler
	}
	return handlers
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
