//go:build ext_test

package spam

import (
	"context"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
)

func init() {
	// Register a resolution
	err := resolutions.RegisterResolution("spam-resolution", resolutions.ResolutionConfig{
		ResolveFunc: func(ctx context.Context, app *common.App, resolution *resolutions.Resolution) error {
			// This is where the resolution logic goes
			app.Service.Logger.Info("Spam resolution logic approved")
			return nil
		},
	})
	if err != nil {
		panic(err)
	}
}
