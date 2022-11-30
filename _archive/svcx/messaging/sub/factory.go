package sub

import (
	"context"
	"fmt"
	"kwil/_archive/svcx/messaging/mx"
	"kwil/x"
	"kwil/x/cfgx"
	"kwil/x/syncx"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
)

func NewChannelBroker(config cfgx.Config) (ChannelBroker, error) {
	cfg, err := mx.NewReceiverConfig(config)
	if err != nil {
		return nil, err
	}

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

func NewTransientReceiver(config cfgx.Config) (TransientReceiver, error) {
	cfg, err := mx.NewReceiverConfig(config)
	if err != nil {
		return nil, err
	}

	if len(cfg.ConsumerTopics) != 1 {
		return nil, fmt.Errorf("transient receiver can only be created for a single topic")
	}

	if cfg.Group != "" {
		return nil, fmt.Errorf("transient receiver cannot be used with a consumer group")
	}

	if len(cfg.AssignPartitions) == 0 {
		return nil, fmt.Errorf("transient receiver must configured with partitions")
	}

	p_map := make(map[int32]kgo.Offset)
	for _, p_id := range cfg.AssignPartitions {
		p_map[p_id] = kgo.NewOffset().AtEnd()
	}

	t_map := make(map[string]map[int32]kgo.Offset)
	t_map[cfg.ConsumerTopics[0]] = p_map

	c, err := kgo.NewClient(
		kgo.ConsumePartitions(t_map),
		kgo.Dialer(cfg.Dialer.DialContext),
		kgo.SASL(plain.Auth{User: cfg.User, Pass: cfg.Pwd}.AsMechanism()),
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.Client_id))

	if err != nil {
		return nil, err
	}

	ctx, fn := context.WithCancel(context.Background())

	return &transient_receiver{
		client:           c,
		topic:            cfg.ConsumerTopics[0],
		out:              make(chan MessageIterator, 32), // todo: buffer should be == to partition count
		done:             make(chan x.Void),
		ctx:              ctx,
		cancelFn:         fn,
		max_poll_records: cfg.MaxPollRecords,
		wg:               &sync.WaitGroup{},
		mu:               &sync.Mutex{},
	}, nil
}
