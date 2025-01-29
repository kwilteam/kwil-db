package orderedsync

import (
	"context"
	"fmt"
	"sync"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

type ResolveFunc func(ctx context.Context, app *common.App, block *common.BlockContext, res *ResolutionMessage) error

// Synchronizer is the global instance of the ordered sync extension.
var Synchronizer = &cachedSync{
	topics: make(map[string]*topicInfo),
}

type topicInfo struct {
	resolve            ResolveFunc
	lastProcessedPoint *int64 // last processed point in time, nil if none have been processed
}

// this file holds a thread-safe in-memory cache for registering topics.
// Unlike extension registration, which happens on init, topic registration
// occurs over the life cycle of the node.

type cachedSync struct {
	// mu protects all fields in this struct
	mu sync.Mutex
	// topics is a map of topics to their respective callbacks
	topics      map[string]*topicInfo
	initialized bool // basic protection against double initialization in case of a bug elsewhere
}

// RegisterTopic registers a topic with a callback.
// It should be called exactly once in the lifecycle of the node
// unless a topic is unregistered.
// (e.g. within a precompile's OnStart method).
func (c *cachedSync) RegisterTopic(ctx context.Context, db sql.DB, eng common.Engine, topic string, cb ResolveFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.topics[topic]
	if ok {
		return fmt.Errorf("topic %s already registered", topic)
	}
	c.topics[topic] = &topicInfo{
		resolve: cb,
	}

	return registerTopic(ctx, db, eng, topic)
}

// UnregisterTopic unregisters a topic.
// It should be called exactly once when a topic is no longer relevant.
// (e.g. within a precompile's OnUnUse method).
func (c *cachedSync) UnregisterTopic(ctx context.Context, db sql.DB, eng common.Engine, topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.topics[topic]
	if !ok {
		return fmt.Errorf("topic %s not registered", topic)
	}

	delete(c.topics, topic)

	return unregisterTopic(ctx, db, eng, topic)
}

// readTopicInfoOnStartup reads the last processed point in time for each topic.
// It is meant to be called within an EngineReadyHook.
// It is called AFTER every topic has been registered.
func (c *cachedSync) readTopicInfoOnStartup(ctx context.Context, app *common.App) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return fmt.Errorf("already initialized, this is an internal error")
	}
	c.initialized = true

	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	SELECT name, last_processed_point FROM topics
	`, nil, func(r *common.Row) error {
		if len(r.Values) != 2 {
			// this should never happen
			return fmt.Errorf("unexpected number of columns")
		}

		point, ok := r.Values[1].(int64)
		if !ok {
			// can be nil, if so, then we just skip
			// because we already have nil as the default
			if r.Values[1] == nil {
				return nil
			}
			return fmt.Errorf("unexpected type int64 for last processed point")
		}

		topic, ok := r.Values[0].(string)
		if !ok {
			return fmt.Errorf("unexpected type string for topic")
		}

		info, ok := c.topics[topic]
		if !ok {
			return fmt.Errorf("data for a topic was found but no topic was registered. topic: %s", topic)
		}

		info.lastProcessedPoint = &point
		return nil
	})
}

// resolve gets all finalized data points from the engine and calls the registered callback.
func (c *cachedSync) resolve(ctx context.Context, app *common.App, block *common.BlockContext) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	res, err := getFinalizedDataPoints(ctx, app.DB, app.Engine)
	if err != nil {
		return err
	}

	// we track the last resolved point so we can
	// update the last processed point in the database
	latestPoint := make(map[string]int64)
	for _, dp := range res {
		top, ok := c.topics[dp.Topic]
		if !ok {
			return fmt.Errorf("topic %s not registered", dp.Topic)
		}

		if err := top.resolve(ctx, app, block, dp); err != nil {
			return err
		}

		latestPoint[dp.Topic] = dp.PointInTime
	}

	// update the last processed point in the database.
	// We do this in order to guarantee deterministic behavior
	for _, kv := range order.OrderMap(latestPoint) {
		err = setLatestPointInTime(ctx, app.DB, app.Engine, kv.Key, kv.Value)
		if err != nil {
			return err
		}

		// update the in-memory cache
		c.topics[kv.Key].lastProcessedPoint = &kv.Value
	}

	return nil
}

// storeDataPoint stores a data point in the engine.
// It ensures that the point and previous point are not less than the previous point in time.
func (c *cachedSync) storeDataPoint(ctx context.Context, db sql.DB, eng common.Engine, dp *ResolutionMessage) error {
	if dp.PreviousPointInTime != nil {
		if dp.PointInTime <= *dp.PreviousPointInTime {
			return fmt.Errorf("point in time must be greater than previous point in time")
		}
	}

	topic, ok := c.topics[dp.Topic]
	if !ok {
		return fmt.Errorf("topic %s not registered", dp.Topic)
	}

	switch {
	case topic.lastProcessedPoint == nil && dp.PreviousPointInTime != nil:
		return fmt.Errorf("non-nil previous point in time received, expected nil because no previous point in time is known")
	case topic.lastProcessedPoint != nil && dp.PreviousPointInTime == nil:
		return fmt.Errorf("nil previous point in time received, expected %d", *topic.lastProcessedPoint)
	case topic.lastProcessedPoint != nil && dp.PreviousPointInTime != nil:
		// they dont have to match, but the incoming must be gte the last processed point
		if *dp.PreviousPointInTime < *topic.lastProcessedPoint {
			return fmt.Errorf("previous point in time must be greater than or equal to the last processed point")
		}
	case topic.lastProcessedPoint == nil && dp.PreviousPointInTime == nil:
		// no previous point in time known, no previous point in time received, all good
	default:
		return fmt.Errorf("unexpected state in storeDataPoint")
	}

	return storeDataPoint(ctx, db, eng, dp)
}
