package migrations

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/extensions/listeners"
)

// Changeset Extension polls the changesets from the old chain  during migration.
// The changesets received are broadcasted to the eventstore under ChangesetMigrationEvent type for voting.
// The changesets are then applied to the new chain's database once the resolution is approved.
const (
	ListenerName = "migrations"
	retryDelay   = 1 * time.Second // polling frequency for changesets
	maxRetries   = 10              // max retries for polling changesets
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
	StartHeight uint64

	// End height is the block height at which the migration ended on the old chain.
	EndHeight uint64

	// ListenAddress is the address of the kwild server to receive changesets from.
	ListenAddress string
}

type migrationListener struct {
	// Config is the migration configuration
	config *MigrationConfig

	// Client is the kwild RPC client to the old chain
	client *client.Client

	// currentHeight is the current block height to be processed
	currentHeight uint64

	// eventStore is the event store to broadcast the changeset migration events
	eventStore listeners.EventStore

	// logger
	logger log.SugaredLogger
}

func Start(ctx context.Context, service *common.Service, eventStore listeners.EventStore) error {
	// Get the migration config from the service
	cfg := &MigrationConfig{}
	migrationCfg, ok := service.LocalConfig.AppCfg.Extensions[ListenerName]
	if !ok {
		service.Logger.Warn("no migration config provided, skipping migration listener")
		return nil // no migration config, nothing to do
	}

	cfg.ListenAddress, ok = migrationCfg["listen_address"]
	if !ok {
		return errors.New("migration listen_address not provided")
	}

	// Kwild RPC client connection to the old chain
	clt, err := client.NewClient(ctx, cfg.ListenAddress, nil)
	if err != nil {
		return err
	}

	cfg.StartHeight = uint64(service.GenesisConfig.ConsensusParams.Migration.StartHeight)
	cfg.EndHeight = uint64(service.GenesisConfig.ConsensusParams.Migration.EndHeight)

	// Get the block height of the last changesets received from the old chain
	// If no height is found, start from the start height of the migrations.
	// If all the changesets have already been received, the listener will exit.
	var currentHeight, lastHeight uint64
	lastHeight, err = getLastStoredHeight(ctx, eventStore)
	if err != nil {
		return err
	}

	if lastHeight == 0 {
		currentHeight = cfg.StartHeight
	} else if lastHeight >= cfg.EndHeight {
		service.Logger.Info("all the changesets have been synced, closing the migrations listener.", log.Int("height", int64(lastHeight)))
		return nil
	} else {
		currentHeight = lastHeight + 1
	}

	// Create the migration listener
	listener := &migrationListener{
		config:        cfg,
		client:        clt,
		currentHeight: currentHeight,
		eventStore:    eventStore,
		logger:        service.Logger,
	}

	// Start polling the admin server for changesets
	service.Logger.Info("start syncing changesets from old chain", log.Int("startHeight", int64(cfg.StartHeight)))
	return listener.RetrieveChangesets(ctx)
}

func (ml *migrationListener) RetrieveChangesets(ctx context.Context) error {
	for {
		if ml.currentHeight >= ml.config.EndHeight {
			// Synced up with the old chain, no more changesets to pull. Close the listener
			ml.logger.Info("changesets have been synchronized with the old chain", log.Int("height", int64(ml.currentHeight)))
			return nil
		}

		// Get the changeset metadata from the admin server
		// retries till the metadata is received successfully using exponential backoff
		// backoff timer is reset on each successful metadata retrieval or after max retries
		numChunks, chunkSizes, err := ml.GetChangesetMetadata(ctx)
		if err != nil {
			continue // Should we return err here?
		}
		btsReceived := int64(0)

		ml.logger.Debug("received changeset metadata", log.Int("height", int64(ml.currentHeight)), log.Int("numChunks", numChunks), log.Any("chunkSizes", chunkSizes))

		wg := sync.WaitGroup{}
		errChan := make(chan error, numChunks)
		for i := int64(0); i < numChunks; i++ {
			wg.Add(1)
			go func(chunkIdx int64) {
				cs, err := ml.LoadChangeset(ctx, chunkIdx, chunkSizes[chunkIdx])
				if err != nil {
					errChan <- err
					return
				}

				// Create the changeset migration event
				totalChunks := uint64(numChunks)
				idx := uint64(chunkIdx)
				evt := &changesetMigration{
					Height:      ml.currentHeight,
					TotalChunks: totalChunks,
					ChunkIdx:    idx,
					Changeset:   cs,
				}

				csEvt, err := evt.MarshalBinary()
				if err != nil {
					ml.logger.Error("failed to marshal changeset migration event", "error", err)
					errChan <- err
					return
				}

				ml.logger.Info("broadcasting changeset migration event", "height", ml.currentHeight, "chunk", chunkIdx, "size", len(cs))

				// Broadcast the changeset migration event to the event store for voting
				err = ml.eventStore.Broadcast(ctx, ChangesetMigrationEventType, csEvt)
				if err != nil {
					errChan <- fmt.Errorf("failed to broadcast changeset migration event: %w", err)
					return
				}

				btsReceived += int64(len(cs))
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

		if err = setLastStoredHeight(ctx, ml.eventStore, ml.currentHeight); err != nil {
			return err
		}
		ml.currentHeight++

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDelay):
		}
	}
}

func (ml *migrationListener) GetChangesetMetadata(ctx context.Context) (totalChunks int64, chunkSizes []int64, err error) {
	err = retry(ctx, maxRetries, func() error {
		totalChunks, chunkSizes, err = ml.client.ChangesetMetadata(ctx, int64(ml.currentHeight))
		return err
	})
	return totalChunks, chunkSizes, err
}

func (ml *migrationListener) LoadChangeset(ctx context.Context, chunkIdx int64, chunkSize int64) (cs []byte, err error) {
	err = retry(ctx, maxRetries, func() error {
		if cs, err = ml.client.LoadChangeset(ctx, int64(ml.currentHeight), chunkIdx); err != nil {
			return err
		}

		if int64(len(cs)) != chunkSize {
			return fmt.Errorf("changeset size mismatch: expected %d, got %d. Maybe try other rpc provider", chunkSize, len(cs))
		}

		return nil
	})
	return cs, err
}

func (c *MigrationConfig) Map() map[string]string {
	return map[string]string{
		"start_height":   strconv.FormatUint(c.StartHeight, 10),
		"end_height":     strconv.FormatUint(c.EndHeight, 10),
		"listen_address": c.ListenAddress,
	}
}

var (
	// lastHeightKey is the key used to store the last height processed by the listener
	lastHeightKey = []byte("lh")
)

// getLastStoredHeight gets the last height stored by the KV store
func getLastStoredHeight(ctx context.Context, eventStore listeners.EventStore) (uint64, error) {
	// get the last confirmed block height processed by the listener
	lastHeight, err := eventStore.Get(ctx, lastHeightKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get last block height: %w", err)
	}

	if len(lastHeight) == 0 {
		return 0, nil
	}

	return binary.LittleEndian.Uint64(lastHeight), nil
}

// setLastStoredHeight sets the last height stored by the KV store
func setLastStoredHeight(ctx context.Context, eventStore listeners.EventStore, height uint64) error {
	heightBts := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBts, height)

	// set the last confirmed block height processed by the listener
	err := eventStore.Set(ctx, lastHeightKey, heightBts)
	if err != nil {
		return fmt.Errorf("failed to set last block height: %w", err)
	}
	return nil
}

// retry will retry the function until it is successful, or reached the max retries
func retry(ctx context.Context, maxRetries int64, fn func() error) error {
	retrier := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    10 * time.Second,
		Factor: 2,
		Jitter: true,
	}

	for {
		err := fn()
		if err == nil {
			return nil
		}

		// fail after maxRetries retries
		if retrier.Attempt() > float64(maxRetries) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retrier.Duration()):
		}
	}
}
