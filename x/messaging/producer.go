package messaging

import (
	"context"
	"fmt"
	"os"
	"sync"

	cfg "kwil/x/messaging/config"
	"kwil/x/rx"

	"github.com/google/uuid"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type Producer[T any] struct {
	kp      *kafka.Producer
	topic   string
	chClose chan struct{}
	serdes  Serdes[T]
	onDone  *rx.Task[struct{}]
	mu      sync.RWMutex
	out     chan *messageWithCtx
}

var ErrProducerClosed = fmt.Errorf("producer closed")
var ErrUnexpectedProducerError = fmt.Errorf("producer event response unknown")

// ToDo: pass *proper* golang style config
func NewProducer[T any](config cfg.Config, serdes Serdes[T]) (*Producer[T], error) {
	if serdes == nil {
		return nil, fmt.Errorf("serdes is nil")
	}

	tp, kp, err := load(config)
	if err != nil {
		return nil, err
	}

	p := &Producer[T]{
		kp:      kp,
		topic:   tp,
		chClose: make(chan struct{}),
		serdes:  serdes,
		onDone:  rx.NewTask[struct{}](),
		mu:      sync.RWMutex{},
		out:     make(chan *messageWithCtx, 100),
	}

	go p.beginEventProcessing(p.onDone, p.out)

	return p, nil
}

func (c *Producer[T]) Submit(ctx context.Context, message T) *rx.Continuation {
	key, payload, err := c.serdes.Serialize(message)
	if err != nil {
		return rx.FailureC(err)
	}

	task := rx.NewTask[struct{}]()

	if ctx == nil {
		ctx = context.Background()
	}

	msg := &messageWithCtx{ctx: ctx, msg: c.createMessage(key, payload, task)}

	c.enqueue(msg)

	return task.AsContinuation()
}

// TODO: look at clean up later -- the close down coordination seems messy
func (c *Producer[T]) Close() {
	c.mu.Lock()
	if c.out == nil {
		return
	}

	out := c.out
	c.out = nil

	c.mu.Unlock()

	close(c.chClose)
	close(out)
}

func (c *Producer[T]) OnClosed() *rx.Continuation {
	return c.onDone.AsContinuation()
}

func (c *Producer[T]) AwaitClosed(ctx context.Context) bool {
	//wait for event loop to complete shutdown
	return c.onDone.Await(ctx)
}

func (c *Producer[T]) enqueue(msg *messageWithCtx) {
	c.mu.RLock()
	e := msg.ctx.Err() //checking in case the lock is held for too long
	if e != nil {
		c.mu.RUnlock()
		msg.fail(e)
		return
	}

	if c.out != nil {
		c.out <- msg
		c.mu.RUnlock()
		return
	}

	if !c.AwaitClosed(msg.ctx) {
		msg.fail(msg.ctx.Err())
	} else {
		msg.fail(ErrProducerClosed)
	}
}

func (c *Producer[T]) createMessage(key []byte, payload []byte, task *rx.Task[struct{}]) *kafka.Message {
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &c.topic},
		Key:            key,
		Value:          payload,
		Opaque:         task,
	}
}

func (c *Producer[T]) doSend(mc *messageWithCtx) {
	if mc.ctx.Err() != nil {
		mc.fail(mc.ctx.Err())
		return
	}

	err := c.kp.Produce(mc.msg, nil)
	if err != nil {
		mc.fail(err)
	}
}

func handleEvent(e kafka.Event, done *int) {
	switch m := e.(type) {
	case *kafka.Message:
		if m.Opaque == nil {
			// Producer is closed or some other unrecoverable error has occured
			// CONFIRM: possible deadlock by CommitBlock() caller if close does not flush out all messages to event that are pending call back
			fmt.Println("Producer is closed or some other unrecoverable error has occured")
			if done != nil {
				*done = 4
			}
			return
		}

		task := m.Opaque.(*rx.Task[struct{}])
		task.CompleteOrFail(struct{}{}, m.TopicPartition.Error)
	}
}

func load(config cfg.Config) (topic string, kp *kafka.Producer, err error) {
	if config == nil {
		return "", nil, fmt.Errorf("config is nil")
	}

	m := make(kafka.ConfigMap)

	settings := config.Select("cluster-settings").ToMap()
	for k, v := range settings {
		m[k] = kafka.ConfigValue(v)
	}

	topic = config.String("topic")
	if topic == "" {
		return "", nil, fmt.Errorf("topic cannot be empty")
	}

	if _, ok := settings["client.id"]; !ok {
		h, _ := os.Hostname()
		m["client.id"] = h + "_" + uuid.New().String()
	}

	m["linger.ms"] = config.Int32("linger-ms", 50)

	p, err := kafka.NewProducer(&m)
	if err != nil {
		return topic, nil, fmt.Errorf("failed to create producer: %s", err)
	}

	return topic, p, nil
}

func (c *Producer[T]) beginEventProcessing(onDone *rx.Task[struct{}], out chan *messageWithCtx) {
	ev := c.kp.Events()
	done := 0
	for done == 0 {
		select {
		case <-c.chClose:
			done = 1
		case m, ok := <-out:
			if !ok {
				done = 2
			} else {
				c.doSend(m)
			}
		case e, ok := <-ev:
			if !ok {
				done = 3
			} else {
				handleEvent(e, &done)
			}
		}
	}

	if done != 1 {
		c.Close()
	}

	c.kp.Close()
	if done < 3 {
		// handle and report remaining messages that have come back
		for e := range ev {
			handleEvent(e, nil)
		}
	}

	if done != 2 {
		for {
			// fail any pending messages
			if m, ok := <-out; ok {
				m.fail(ErrProducerClosed)
				continue
			}

			break
		}
	}

	if done == 4 {
		onDone.Fail(ErrUnexpectedProducerError)
	} else {
		onDone.Complete(struct{}{})
	}
}

type messageWithCtx struct {
	ctx context.Context
	msg *kafka.Message
}

func (c *messageWithCtx) fail(err error) {
	c.msg.Opaque.(*rx.Task[struct{}]).Fail(err)
}
