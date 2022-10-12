package sub

//
//import "C"
//import (
//	"context"
//	"fmt"
//	"github.com/twmb/franz-go/pkg/kgo"
//	"github.com/twmb/franz-go/pkg/sasl/plain"
//	"kwil/x"
//	"kwil/x/messaging/mx"
//	"kwil/x/syncx"
//	"sync"
//)
//
//type channel_broker[T any] struct {
//	consumer          *kgo.Client
//	consumer_factory  func() (*kgo.Client, error)
//	serdes            mx.Serdes[T]
//	receiver_assigned chan ReceiverChannel[T]
//	done              chan x.Void
//	pending           sync.WaitGroup
//	mu                sync.Mutex
//	shutdown          syncx.Chan[x.Void]
//	receivers         map[string]map[int32]ReceiverChannel[T]
//	ctx               context.Context
//	cancelFn          context.CancelFunc
//	max_poll_records  int
//}
//
//func new_channel_broker[T any](cfg *mx.ClientConfig[T], serdes mx.Serdes[T]) (*channel_broker[T], error) {
//	cb := &channel_broker[T]{
//		serdes:            cfg.Serdes,
//		receiver_assigned: make(chan ReceiverChannel[T], 32),
//		done:              make(chan x.Void),
//		pending:           sync.WaitGroup{},
//		mu:                sync.Mutex{},
//		shutdown:          syncx.NewChan[x.Void](),
//		receivers:         make(map[string]map[int32]ReceiverChannel[T]),
//		max_poll_records:  100,
//	}
//
//	c, err := kgo.NewClient(
//		kgo.SeedBrokers(cfg.Brokers...),
//		kgo.ProducerLinger(cfg.Linger),
//		kgo.ClientID(cfg.Client_id),
//		kgo.SASL(plain.Auth{User: cfg.User, Pass: cfg.Pwd}.AsMechanism()),
//		kgo.Dialer(cfg.Dialer.DialContext),
//		kgo.ConsumerGroup(cfg.Group),
//		kgo.OnPartitionsAssigned(cb.handlePartitionsAssigned),
//		kgo.OnPartitionsRevoked(cb.handlePartitionsRevoked),
//	)
//
//	if err != nil {
//		return nil, err
//	}
//
//	ctx, fn := context.WithCancel(context.Background())
//	cb.ctx = ctx
//	cb.cancelFn = fn
//	cb.consumer = c
//
//	return cb, nil
//}
//
//func (b *channel_broker[T]) Start(topics ...string) error {
//	b.mu.Lock()
//	if b.consumer != nil {
//		return fmt.Errorf("broker already started")
//	}
//
//	c, err := b.consumer_factory()
//	if err != nil {
//		return fmt.Errorf("failed to create consumer: %s", err)
//	}
//
//	b.consumer = c
//	b.mu.Unlock()
//
//	go func() {
//		for !b.shutdown.IsClosed() {
//			fetches := c.PollRecords(b.ctx, b.max_poll_records)
//			err := fetches.Err()
//			if err != nil {
//				b.Shutdown()
//				return
//			}
//
//			// get receivers and emit records
//			//for _, record := range fetches.Records() {
//			//	b.pending.Add(1)
//			//
//			//	receivers := b.getReceivers(record.Topic, record.Partition)
//			//	go func() {
//			//		defer b.pending.Done()
//			//		b.handleRecord(record)
//			//	}()
//			//}
//		}
//	}()
//
//	return nil
//}
//
//func (b *channel_broker[T]) handlePartitionsAssigned(ctx context.Context, _ *kgo.Client, partitions map[string][]int32) {
//	b.mu.Lock()
//	if b.shutdown.IsClosed() {
//		b.mu.Unlock()
//		return
//	}
//
//	var channels []ReceiverChannel[T]
//	for t, pt := range partitions {
//		r := b.receivers[t]
//		if r == nil {
//			r = make(map[int32]ReceiverChannel[T])
//			b.receivers[t] = r
//		}
//
//		for _, p := range pt {
//			c := r[p]
//			if c != nil {
//				fmt.Printf("already assigned - topic: %s, partititon: %d\n", t, p)
//				continue
//			} else {
//				c = b.getReceiverChannel()
//				r[p] = c
//				channels = append(channels, c)
//				fmt.Printf("assigned - topic: %s, partititon: %d\n", t, p)
//			}
//			channels = append(channels, c)
//		}
//	}
//
//	if len(channels) == 0 {
//		b.mu.Unlock()
//		return
//	}
//
//	b.pending.Add(1)
//	b.mu.Unlock()
//
//	go func() {
//		for _, c := range channels {
//			b.receiver_assigned <- c
//		}
//		b.pending.Done()
//	}()
//}
//
//func (b *channel_broker[T]) getReceiverChannel() ReceiverChannel[T] {
//	panic("implement me")
//}
//
//func (b *channel_broker[T]) handlePartitionsRevoked(ctx context.Context, _ *kgo.Client, partitions map[string][]int32) {
//	if ctx.Err() != nil {
//		return
//	}
//
//	b.mu.Lock()
//	if b.shutdown.IsClosed() {
//		b.mu.Unlock()
//		return
//	}
//
//	var channels []ReceiverChannel[T]
//	for t, pt := range partitions {
//		t_map := b.receivers[t]
//		if t_map == nil {
//			continue
//		}
//
//		for _, p := range pt {
//			c := t_map[p]
//			if c != nil {
//				channels = append(channels, c)
//				delete(t_map, p)
//				fmt.Printf("revoked - topic: %s, partititon: %d\n", t, p)
//			}
//		}
//	}
//
//	if len(channels) == 0 {
//		b.mu.Unlock()
//		return
//	}
//
//	b.pending.Add(1)
//
//	go func() {
//		b.clean_up_receivers(b.receivers)
//		b.pending.Done()
//	}()
//}
//
//func (b *channel_broker[T]) OnChannelAssigned() <-chan ReceiverChannel[T] {
//	return b.receiver_assigned
//}
//
//func (b *channel_broker[T]) Shutdown() {
//	b.mu.Lock()
//	if !b.shutdown.Close() {
//		return
//	}
//
//	b.mu.Unlock()
//
//	b.pending.Wait() // wait for pending operations
//
//	receivers := b.receivers
//	b.receivers = nil
//
//	b.mu.Lock()
//
//	shutdown := b.shutdown
//	b.shutdown = syncx.ClosedChanVoid()
//
//	b.mu.Unlock()
//
//	shutdown.Close()
//
//	if b.consumer == nil {
//		// never started, close and return
//		close(b.done)
//		return
//	}
//
//	b.cancelFn()
//
//	go func() {
//		b.clean_up_receivers(receivers)
//		// b.consumer.CloseAllowingRebalance() // if we use BlockRebalanceOnPoll
//		b.consumer.Close()
//
//		close(b.done)
//	}()
//}
//
//func (b *channel_broker[T]) clean_up_receivers(receivers map[string]map[int32]ReceiverChannel[T]) {
//	for _, subs := range receivers {
//		for _, sub := range subs {
//			sub.Close()
//		}
//	}
//
//	for _, subs := range receivers {
//		for _, sub := range subs {
//			<-sub.OnClosed()
//		}
//	}
//}
//
//func (b *channel_broker[T]) ShutdownAndWait(ctx context.Context) error {
//	b.Shutdown()
//	select {
//	case <-b.done:
//		return nil
//	case <-ctx.Done():
//		return ctx.Err()
//	}
//}
//
//func (b *channel_broker[T]) OnShutdown() <-chan x.Void {
//	return b.done
//}
