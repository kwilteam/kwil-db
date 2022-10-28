package sub

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"kwil/x"
	"kwil/x/async"
	"kwil/x/svcx/messaging/mx"
	"math"
	"sync"
)

type transient_receiver struct {
	client           *kgo.Client
	topic            string
	out              chan MessageIterator
	done             chan x.Void
	ctx              context.Context
	cancelFn         context.CancelFunc
	max_poll_records int
	wg               *sync.WaitGroup
	mu               *sync.Mutex
	started          bool
}

func (c *transient_receiver) Topic() string {
	return c.topic
}

func (c *transient_receiver) OnReceive() <-chan MessageIterator {
	return c.out
}

// TODO: look at possible need to start at offset for partitions (depending on usage, it may be a non issue)
func (c *transient_receiver) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("receiver already started")
	}

	c.started = true

	go func() {
		for {
			f := c.client.PollRecords(c.ctx, c.max_poll_records)
			err := f.Err()
			if err == nil {
				c._process(f)
				continue
			}

			if err != context.Canceled {
				// TODO: use logger here
				fmt.Printf("topic (%s) consumer error: %s", c.topic, err)
			}

			break
		}

		close(c.out)
		c.client.Close()
		close(c.done)
	}()

	return nil
}

func (c *transient_receiver) Stop() {
	c.cancelFn() // safe to call multiple times
}

func (c *transient_receiver) OnStop() <-chan x.Void {
	return c.done
}

func (c *transient_receiver) _process(fetches kgo.Fetches) {
	var partitions []*x.Tuple2[mx.PartitionId, []*kgo.Record]
	fetches.EachPartition(func(ftp kgo.FetchTopicPartition) {
		var records []*kgo.Record
		ftp.EachRecord(func(r *kgo.Record) {
			records = append(records, r)
		})

		p_id := mx.PartitionId(ftp.Partition)
		partitions = append(partitions, x.NewTuple2(p_id, records))
	})

	if len(partitions) == 0 {
		return
	}

	c.wg.Add(len(partitions))

	for _, p := range partitions {
		select {
		case <-c.ctx.Done():
			break
		default:
			if !c.enqueue(p) {
				break
			}
		}
	}

	c.wg.Wait()
}

func (c *transient_receiver) enqueue(p *x.Tuple2[mx.PartitionId, []*kgo.Record]) bool {
	once := &sync.Once{}
	wg_done := func() {
		once.Do(func() {
			c.wg.Done()
		})
	}

	index := -1
	next := func() (msg *mx.RawMessage, offset mx.Offset) {
		index++
		if index >= len(p.P2) || c.ctx.Err() != nil {
			wg_done()
			return nil, mx.Offset(math.MinInt)
		}

		r := p.P2[index]
		return &mx.RawMessage{Key: r.Key, Value: r.Value}, mx.Offset(r.Offset)
	}

	iter := &message_iterator{p.P1, next, c.getCommitFn(wg_done), nil, -1}

	c.wg.Add(1)

	select {
	case c.out <- iter:
		return true
	case <-c.ctx.Done():
		wg_done()
		return false
	}
}

func (c *transient_receiver) getCommitFn(fn func()) func() async.Action {
	return func() async.Action {
		fn()
		return async.CompletedAction()
	}
}
