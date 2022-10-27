package sub

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"kwil/x"
	"kwil/x/svcx/messaging/mx"
	"kwil/x/syncx"
	"sync"
)

func NewChannelBroker(cfg *mx.ClientConfig) (ChannelBroker, error) {
	cb := &channel_broker{
		receiver_assigned: make(chan ReceiverChannel, 32),
		done:              make(chan x.Void),
		pending:           sync.WaitGroup{},
		mu:                sync.Mutex{},
		shutdown:          syncx.NewChan[x.Void](),
		receivers:         make(map[string]map[int32]ReceiverChannel),
		max_poll_records:  100,
	}

	if len(cfg.ConsumerTopics) == 0 {
		return nil, fmt.Errorf("no topics configured")
	}

	if cfg.Group == "" {
		return nil, fmt.Errorf("group is required")
	}

	c, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.Client_id),
		kgo.SASL(plain.Auth{User: cfg.User, Pass: cfg.Pwd}.AsMechanism()),
		kgo.Dialer(cfg.Dialer.DialContext),
		kgo.ConsumeTopics(cfg.ConsumerTopics...),
		kgo.AutoCommitMarks(),
		kgo.ConsumerGroup(cfg.Group),
		kgo.OnPartitionsAssigned(cb.handlePartitionsAssigned),
		kgo.OnPartitionsRevoked(cb.handlePartitionsRevoked),
	)

	if err != nil {
		return nil, err
	}

	ctx, fn := context.WithCancel(context.Background())
	cb.ctx = ctx
	cb.cancelFn = fn
	cb.consumer = c

	return cb, nil
}
func NewTransientReceiver(cfg *mx.ClientConfig) (TransientReceiver, error) {
	if len(cfg.ConsumerTopics) != 1 {
		return nil, fmt.Errorf("transient receiver can only be created for a single topic")
	}

	if cfg.Group != "" {
		return nil, fmt.Errorf("transient receiver cannot be used with a consumer group")
	}

	c, err := kgo.NewClient(
		kgo.Dialer(cfg.Dialer.DialContext),
		kgo.SASL(plain.Auth{User: cfg.User, Pass: cfg.Pwd}.AsMechanism()),
		kgo.ConsumeTopics(cfg.ConsumerTopics...),
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.Client_id),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()))

	if err != nil {
		return nil, err
	}

	ctx, fn := context.WithCancel(context.Background())

	return &transient_receiver{
		c,
		cfg.ConsumerTopics[0],
		syncx.NewChanBuffered[MessageIterator](1),
		make(chan x.Void),
		ctx,
		fn,
		cfg.MaxPollRecords,
		sync.Mutex{},
		false,
	}, nil
}
