package setup

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/testcontainers/testcontainers-go"
)

type AdminClient struct {
	container *testcontainers.DockerContainer
}

// exec runs a command in the admin container.
// It will JSON unmarshal the result into result.
func exec[T any](a *AdminClient, ctx context.Context, result T, args ...string) error {
	_, reader, err := a.container.Exec(ctx, append([]string{"/app/kwil-admin"}, args...))
	if err != nil {
		return err
	}

	d := display.MessageReader[T]{
		Result: result,
	}

	err = json.NewDecoder(reader).Decode(&d)
	if err != nil {
		return err
	}

	if d.Error != "" {
		return errors.New(d.Error)
	}

	return nil
}
func (a *AdminClient) ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte) (types.Hash, error) {
	var hash types.Hash
	err := exec(a, ctx, &hash, "validators", "approve", string(joinerPubKey))
	return hash, err
}
