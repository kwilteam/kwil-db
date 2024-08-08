package migrations

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/adminclient"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/extensions/listeners"
)

// Changeset Extension polls the changesets from the old chain  during migration.
// The changesets received are broadcasted to the eventstore under ChangesetMigrationEvent type for voting.
// The changesets are then applied to the new chain's database once the resolution is approved.
const (
	ListenerName = "migrations"
	retryDelay   = 1 * time.Second // polling frequency for changesets
)

func init() {
	// Register the listener with the name "migrations"
	err := listeners.RegisterListener(ListenerName, Start)
	if err != nil {
		panic(err)
	}
}

type MigrationConfig struct {
	// StartHeight is the block height at which the migration started on the old chain.
	StartHeight int64

	// End height is the block height at which the migration ended on the old chain.
	EndHeight int64

	// AdminListenAddress is the address of the admin server to receive changesets from.
	AdminListenAddress string

	// AdminPass is the admin server password.
	AdminPass string

	// KwildTLSCertFile is the path to the TLS certificate file for the Kwil node.
	// The path should be an absolute path or relative to the node's root directory.
	KwildTLSCertFile string
}

func Start(ctx context.Context, service *common.Service, eventStore listeners.EventStore) error {
	// Get the migration config from the service
	cfg := &MigrationConfig{}
	migrationCfg, ok := service.ExtensionConfigs[ListenerName]
	if !ok {
		service.Logger.Warn("no migration config provided, skipping migration listener")
		return nil // no migration config, nothing to do
	}

	err := cfg.ExtractConfig(migrationCfg)
	if err != nil {
		return err
	}

	var dialOpt []adminclient.Opt
	if cfg.KwildTLSCertFile != "" {
		dialOpt = append(dialOpt, adminclient.WithTLS(cfg.KwildTLSCertFile, "", ""))
	}

	if cfg.AdminPass != "" {
		dialOpt = append(dialOpt, adminclient.WithPass(cfg.AdminPass))
	}

	// Admin RPC client connection to the old chain
	adminClient, err := adminclient.NewClient(ctx, cfg.AdminListenAddress, dialOpt...)
	if err != nil {
		return err
	}

	// Get the block height of the last changesets received from the old chain
	// If no height is found, start from the start height of the migrations.
	// If all the changesets have already been received, the listener will exit.
	var currentHeight, lastHeight int64
	lastHeight, err = getLastStoredHeight(ctx, eventStore)
	if err != nil {
		return err
	}

	if lastHeight == 0 {
		currentHeight = cfg.StartHeight
	} else if lastHeight >= cfg.EndHeight {
		service.Logger.Info("all the changesets have been synced to the new chain, closing the migrations listener.", log.Int("height", lastHeight))
		return nil
	} else {
		currentHeight = lastHeight + 1
	}

	// Start polling the admin server for changesets
	service.Logger.Info("start syncing changesets from old chain", log.Int("startHeight", cfg.StartHeight))
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if currentHeight >= cfg.EndHeight {
				// Synced up with the old chain, no more changesets to pull. Close the listener
				service.Logger.Info("changesets have been synchronized with the old chain", log.Int("height", currentHeight))
				return nil
			}

			// Get the changeset metadata from the admin server
			numChunks, size, err := adminClient.ChangesetMetadata(ctx, currentHeight)
			if err != nil {
				fmt.Println("Couldnt fetch changesets", err)
				time.Sleep(retryDelay) // If no changeset is found, wait and try again
				continue
			}
			btsReceived := int64(0)

			service.Logger.Debug("received changeset metadata", log.Int("height", currentHeight), log.Int("numChunks", numChunks), log.Int("size", size))
			wg := sync.WaitGroup{}
			errChan := make(chan error, numChunks)

			for i := int64(0); i < numChunks; i++ {
				wg.Add(1)
				go func(chunkIdx int64) {
					for {
						// Load the changeset from the admin server
						cs, err := adminClient.LoadChangeset(ctx, currentHeight, chunkIdx)
						if err != nil {
							// If no changeset is found, wait and try again
							time.Sleep(retryDelay)
							continue
						}

						// Create the changeset migration event
						evt := &ChangesetMigration{
							Height:      big.NewInt(currentHeight),
							TotalChunks: big.NewInt(numChunks),
							ChunkIdx:    big.NewInt(chunkIdx),
							Changeset:   cs,
						}

						csEvt, err := evt.MarshalBinary()
						if err != nil {
							service.Logger.Error("failed to marshal changeset migration event", log.String("error", err.Error()))
							errChan <- err
							return
						}

						service.Logger.Info("broadcasting changeset migration event", log.Int("height", currentHeight), log.Int("chunk", chunkIdx), log.Int("size", len(cs)))

						// Broadcast the changeset migration event to the event store for voting
						err = eventStore.Broadcast(ctx, ChangesetMigrationEventType, csEvt)
						if err != nil {
							errChan <- fmt.Errorf("failed to broadcast changeset migration event: %w", err)
							return
						}

						btsReceived += int64(len(cs))
						break
					}
					wg.Done()
				}(i)
			}

			// Wait for all chunks to be successfully received
			wg.Wait()
			close(errChan)

			// check errChan for all the errors
			var errs error
			for err := range errChan {
				errs = errors.Join(errs, err)
			}

			if errs != nil {
				return errs
			}

			if btsReceived != size {
				return fmt.Errorf("incorrect changeset size received: %d, expected: %d", btsReceived, size)
			}

			if err = setLastStoredHeight(ctx, eventStore, currentHeight); err != nil {
				return err
			}
			currentHeight++

			// Delay before polling the admin server again for the next changeset
			time.Sleep(retryDelay)
		}
	}
}

func (c *MigrationConfig) ExtractConfig(cfg map[string]string) error {
	var err error
	var ok bool

	startHeight, ok := cfg["start_height"]
	if !ok {
		return errors.New("migration start_height not provided")
	}
	c.StartHeight, err = strconv.ParseInt(startHeight, 10, 64)
	if err != nil {
		return err
	}

	endHeight, ok := cfg["end_height"]
	if !ok {
		return errors.New("migration end_height not provided")
	}
	c.EndHeight, err = strconv.ParseInt(endHeight, 10, 64)
	if err != nil {
		return err
	}

	c.AdminListenAddress, ok = cfg["admin_listen_address"]
	if !ok {
		return errors.New("migration admin_listen_address not provided")
	}

	c.AdminPass = cfg["admin_pass"]
	c.KwildTLSCertFile = cfg["kwild_tls_cert_file"]

	return nil
}

func (c *MigrationConfig) Map() map[string]string {
	return map[string]string{
		"start_height":         strconv.FormatInt(c.StartHeight, 10),
		"end_height":           strconv.FormatInt(c.EndHeight, 10),
		"admin_listen_address": c.AdminListenAddress,
		"admin_pass":           c.AdminPass,
		"kwild_tls_cert_file":  c.KwildTLSCertFile,
	}
}

var (
	// lastHeightKey is the key used to store the last height processed by the listener
	lastHeightKey = []byte("lh")
)

// getLastStoredHeight gets the last height stored by the KV store
func getLastStoredHeight(ctx context.Context, eventStore listeners.EventStore) (int64, error) {
	// get the last confirmed block height processed by the listener
	lastHeight, err := eventStore.Get(ctx, lastHeightKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get last block height: %w", err)
	}

	if len(lastHeight) == 0 {
		return 0, nil
	}

	return int64(binary.LittleEndian.Uint64(lastHeight)), nil
}

// setLastStoredHeight sets the last height stored by the KV store
func setLastStoredHeight(ctx context.Context, eventStore listeners.EventStore, height int64) error {
	heightBts := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBts, uint64(height))

	// set the last confirmed block height processed by the listener
	err := eventStore.Set(ctx, lastHeightKey, heightBts)
	if err != nil {
		return fmt.Errorf("failed to set last block height: %w", err)
	}
	return nil
}
