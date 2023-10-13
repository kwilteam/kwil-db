package engine

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/engine/dataset"
)

type extensionInitializerAdapter struct {
	ExtensionInitializer
}

func (e extensionInitializerAdapter) Initialize(ctx context.Context, meta map[string]string) (dataset.InitializedExtension, error) {
	return e.ExtensionInitializer.CreateInstance(ctx, meta)
}
