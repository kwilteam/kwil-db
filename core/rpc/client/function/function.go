// package function specifies the client interface for Kwil's function service.
package function

import (
	"context"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
)

type FunctionServiceClient interface {
	VerifySignature(ctx context.Context, sender []byte, signature *auth.Signature, message []byte) error
}
