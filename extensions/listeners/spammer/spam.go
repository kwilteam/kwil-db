//go:build ext_test

package spammer

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/extensions/listeners"
)

const ListenerName = "spammer"

func init() {
	err := listeners.RegisterListener(ListenerName, Start)
	if err != nil {
		panic(err)
	}
}

func Start(ctx context.Context, service *common.Service, eventStore listeners.EventStore) error {
	if _, ok := service.ExtensionConfigs[ListenerName]; !ok {
		return fmt.Errorf("spammer listener not configured, not starting spam oracle")
	}
	service.Logger.Info("Starting spam oracle")
	count := 0
	for {
		select {
		case <-ctx.Done(): // Properly handle context cancellation
			return nil
		default:
			err := eventStore.Broadcast(ctx, "spam-resolution", []byte(fmt.Sprintf("spam%d", count)))
			if err != nil {
				return err
			}
			count++
			if count == 10000 {
				return nil
			}
		}
	}
}
