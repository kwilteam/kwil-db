package spam

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/extensions/oracles"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
)

func init() {
	fmt.Println("spam oracle registered")
	err := oracles.RegisterOracle("spam", func(ctx context.Context, service *common.Service, eventstore oracles.EventStore) {
		count := 0
		return
		for {
			select {
			case <-ctx.Done():
				fmt.Println("spam oracle stopped")
				return
			default:
				err := eventstore.Broadcast(ctx, "spam", []byte("spam"+fmt.Sprint(count)))
				if err != nil {
					panic(err)
				}

				count++

				if count == 1000000 {
					return
				}
			}
		}
	})
	if err != nil {
		panic(err)
	}

	err = resolutions.RegisterResolution("spam", resolutions.ResolutionConfig{
		ResolveFunc: func(ctx context.Context, app *common.App, resolution *resolutions.Resolution) error {
			fmt.Println("spam resolution resolved")
			return nil
		},
	})
	if err != nil {
		panic(err)
	}
}
