package sub

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"kwil/x"
	"kwil/x/syncx"
	"sync"
)

type transient_receiver struct {
	client           *kgo.Client
	topic            string
	out              syncx.Chan[MessageIterator]
	done             chan x.Void
	ctx              context.Context
	cancelFn         context.CancelFunc
	max_poll_records int
	mu               sync.Mutex
	started          bool
}

func (c *transient_receiver) Topic() string {
	return c.topic
}

func (c *transient_receiver) OnReceive() <-chan MessageIterator {
	return c.out.Read()
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
				c._send(f)
				continue
			}

			if err != context.Canceled {
				// TODO: use logger here
				fmt.Printf("topic (%s) consumer error: %s", c.topic, err)
			}

			break
		}

		c.client.Close()
		c.out.Close()

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

func (c *transient_receiver) _send(fetches kgo.Fetches) {
	// todo - get records into message iterator and write to channel
	//fetches.EachPartition(func(ftp kgo.FetchTopicPartition) {
	//	var records []*kgo.Record
	//	ftp.EachRecord(func(r *kgo.Record) {
	//		records = append(records, r)
	//	})
	//
	//	c.out.Write()
	//})
}
