package sub

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"kwil/x"
	"kwil/x/syncx"
	"sync"
)

type channel_broker struct {
	consumer          *kgo.Client
	consumer_factory  func() (*kgo.Client, error)
	receiver_assigned chan ReceiverChannel
	done              chan x.Void
	pending           sync.WaitGroup
	mu                sync.Mutex
	shutdown          syncx.Chan[x.Void]
	receivers         map[string]map[int32]ReceiverChannel
	ctx               context.Context
	cancelFn          context.CancelFunc
	max_poll_records  int
}

func (b *channel_broker) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.consumer != nil {
		return fmt.Errorf("broker already started")
	}

	if b.shutdown.IsClosed() {
		return fmt.Errorf("broker already stopped")
	}

	c, err := b.consumer_factory()
	if err != nil {
		return fmt.Errorf("failed to create consumer: %s", err)
	}

	b.consumer = c

	go func() {
		for !b.shutdown.IsClosed() {
			fetches := c.PollRecords(b.ctx, b.max_poll_records)
			err := fetches.Err()
			if err != nil {
				b.Stop()
				return
			}

			// get receivers and emit records
			//for _, record := range fetches.Records() {
			//	b.pending.Add(1)
			//
			//	receivers := b.getReceivers(record.Topic, record.Partition)
			//	go func() {
			//		defer b.pending.Done()
			//		b.handleRecord(record)
			//	}()
			//}
		}
	}()

	return nil
}

func (b *channel_broker) OnChannelAssigned() <-chan ReceiverChannel {
	return b.receiver_assigned
}

func (b *channel_broker) Stop() {
	b.mu.Lock()
	if !b.shutdown.Close() {
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()

	b.pending.Wait() // wait for pending operations

	b.mu.Lock()
	receivers := b.receivers
	b.receivers = nil
	shutdown := b.shutdown
	b.shutdown = syncx.ClosedChanVoid()
	b.mu.Unlock()

	shutdown.Close()

	if b.consumer == nil {
		// never started, close and return
		close(b.done)
		return
	}

	b.cancelFn()

	go func() {
		b.clean_up_receivers(receivers)
		// b.consumer.CloseAllowingRebalance() // if we use BlockRebalanceOnPoll
		b.consumer.Close()

		close(b.done)
	}()
}

func (b *channel_broker) OnStop() <-chan x.Void {
	return b.done
}

func (b *channel_broker) handlePartitionsAssigned(ctx context.Context, _ *kgo.Client, partitions map[string][]int32) {
	if ctx.Err() != nil {
		return
	}

	b.mu.Lock()
	if b.shutdown.IsClosed() {
		b.mu.Unlock()
		return
	}

	var channels []ReceiverChannel
	for t, pt := range partitions {
		r := b.receivers[t]
		if r == nil {
			r = make(map[int32]ReceiverChannel)
			b.receivers[t] = r
		}

		for _, p := range pt {
			c := r[p]
			if c != nil {
				fmt.Printf("already assigned - topic: %s, partititon: %d\n", t, p)
				continue
			} else {
				c = b.getReceiverChannel()
				r[p] = c
				channels = append(channels, c)
				fmt.Printf("assigned - topic: %s, partititon: %d\n", t, p)
			}
			channels = append(channels, c)
		}
	}

	if len(channels) == 0 {
		b.mu.Unlock()
		return
	}

	b.pending.Add(1)
	b.mu.Unlock()

	go func() {
		for _, c := range channels {
			b.receiver_assigned <- c
		}
		b.pending.Done()
	}()
}

func (b *channel_broker) getReceiverChannel() ReceiverChannel {
	panic("implement me")
}

func (b *channel_broker) handlePartitionsRevoked(ctx context.Context, _ *kgo.Client, partitions map[string][]int32) {
	if ctx.Err() != nil {
		return
	}

	b.mu.Lock()
	if b.shutdown.IsClosed() {
		b.mu.Unlock()
		return
	}

	var channels []ReceiverChannel
	for t, pt := range partitions {
		t_map := b.receivers[t]
		if t_map == nil {
			continue
		}

		for _, p := range pt {
			c := t_map[p]
			if c != nil {
				channels = append(channels, c)
				delete(t_map, p)
				fmt.Printf("revoked - topic: %s, partititon: %d\n", t, p)
			}
		}
	}

	if len(channels) == 0 {
		b.mu.Unlock()
		return
	}

	b.pending.Add(1)
	b.mu.Unlock()

	go func() {
		for _, c := range channels {
			c.Stop()
		}

		for _, c := range channels {
			<-c.OnStop()
		}

		b.pending.Done()
	}()
}

func (b *channel_broker) clean_up_receivers(receivers map[string]map[int32]ReceiverChannel) {
	for _, subs := range receivers {
		for _, sub := range subs {
			sub.Stop()
		}
	}

	for _, subs := range receivers {
		for _, sub := range subs {
			<-sub.OnStop()
		}
	}
}
