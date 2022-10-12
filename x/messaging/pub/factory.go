package pub

import (
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"kwil/x"
	"kwil/x/cfgx"
	"kwil/x/messaging/internal"
	"kwil/x/messaging/mx"
	"kwil/x/syncx"
	"sync"
	"sync/atomic"
)

// NewEmitterSingleClient creates a new emitter that uses a single client.
// The client is created using the provided config. A client that connects to
// a cluster with multiple brokers will multiplex the emitter to all brokers.
func NewEmitterSingleClient[T any](config cfgx.Config, serdes mx.Serdes[T]) (Emitter[T], error) {
	err := assertValid(serdes)
	if err != nil {
		return nil, err
	}

	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	return NewEmitter(client, serdes)
}

var counter int32

// NewEmitter creates a new emitter that uses the provided client.
func NewEmitter[T any](client mx.Client, serdes mx.Serdes[T]) (Emitter[T], error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	err := assertValid(serdes)
	if err != nil {
		return nil, err
	}

	c, ok := client.(*emitter_client)
	if !ok {
		return nil, fmt.Errorf("invalid client type")
	}

	e := &emitter[T]{
		id:     int(atomic.AddInt32(&counter, 1)),
		client: c,
		serdes: serdes,
	}

	fn, err := c.attach(e)
	if err != nil {
		return nil, err
	}

	e.fn = fn

	return e, nil
}

// NewClient creates a new client that uses the provided config. A client
// that connects to a cluster is used to multiplex using a single underlying
// producer. Once all connect emitters are closed, the client will be closed.
// Conversely, if the client is closed, all emitters will be closed.
func NewClient(config cfgx.Config) (mx.Client, error) {
	cfg, err := mx.NewEmitterClientConfig(config)
	if err != nil {
		return nil, err
	}

	var out syncx.Chan[*message_with_ctx]
	if cfg.Buffer == 0 {
		out = syncx.NewChan[*message_with_ctx]()
	} else {
		out = syncx.NewChanBuffered[*message_with_ctx](cfg.Buffer)
	}

	kp, err := kgo.NewClient(
		kgo.DefaultProduceTopic(cfg.DefaultTopic),
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ProducerLinger(cfg.Linger),
		kgo.ClientID(cfg.Client_id),
		kgo.SASL(plain.Auth{User: cfg.User, Pass: cfg.Pwd}.AsMechanism()),
		kgo.Dialer(cfg.Dialer.DialContext),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create emitter: %s", err)
	}

	e := &emitter_client{
		kp:       kp,
		out:      out,
		done:     syncx.NewChan[x.Void](),
		mu:       sync.Mutex{},
		emitters: make(map[int]internal.Closable),
	}

	go e.begin_processing()

	return e, nil
}

func assertValid[T any](serdes mx.Serdes[T]) error {
	if serdes == nil {
		return fmt.Errorf("serdes is nil")
	}

	return nil
}
