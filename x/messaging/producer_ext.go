package messaging

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	cfg "kwil/x/messaging/config"
	"kwil/x/rx"
	"kwil/x/syncx"
	"kwil/x/utils"
	"os"
)

type producer[T any] struct {
	kp     *kafka.Producer
	topic  string
	serdes Serdes[T]
	out    syncx.Chan[*messageWithCtx]
	done   chan rx.Void
}

func (p *producer[T]) Submit(ctx context.Context, message T) *rx.Continuation {
	key, payload, err := p.serdes.Serialize(message)
	if err != nil {
		return rx.FailureC(err)
	}

	task := rx.NewTask[rx.Void]()

	msg := &messageWithCtx{
		ctx: utils.IfElse(ctx != nil, ctx, context.Background()),
		msg: p.createMessage(key, payload, task),
	}

	if !p.out.Write(msg) {
		task.Fail(ErrProducerClosed)
	}

	return task.AsContinuation()
}

func (p *producer[T]) Close() {
	p.out.Close()
}

func (p *producer[T]) OnClosed() <-chan rx.Void {
	return p.done
}

func (p *producer[T]) createMessage(key []byte, payload []byte, task rx.Task[rx.Void]) *kafka.Message {
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic},
		Key:            key,
		Value:          payload,
		Opaque:         task,
	}
}

func (p *producer[T]) doSend(mc *messageWithCtx) {
	if mc.ctx.Err() != nil {
		mc.fail(mc.ctx.Err())
		return
	}

	err := p.kp.Produce(mc.msg, nil)
	if err != nil {
		mc.fail(err)
	}
}

func handleEvent(e kafka.Event, done *int) {
	switch m := e.(type) {
	case *kafka.Message:
		if m.Opaque == nil {
			// producer is closed or some other unrecoverable error has occurred
			// CONFIRM: possible deadlock by CommitBlock() caller if close does not flush out all messages to event that are pending call back
			fmt.Println("producer is closed or some other unrecoverable error has occurred")
			if done != nil {
				*done = 3
			}
			return
		}

		task := m.Opaque.(rx.Task[rx.Void])
		task.CompleteOrFail(rx.Void{}, m.TopicPartition.Error)
	}
}

func (p *producer[T]) beginEventProcessing() {
	ev := p.kp.Events()
	done := 0
	for done == 0 {
		select {
		case <-p.out.ClosedCh():
			done = 1
		case m, ok := <-p.out.Read():
			if !ok {
				done = 1
			} else {
				p.doSend(m)
			}
		case e, ok := <-ev:
			if !ok {
				done = 2
			} else {
				handleEvent(e, &done)
			}
		}
	}

	if done != 1 {
		p.Close()
	}

	p.kp.Close()
	if done < 3 {
		// handle and report remaining messages that have come back
		for e := range ev {
			handleEvent(e, nil)
		}
	}

	err := utils.IfElse(done == 3, ErrUnexpectedProducerError, ErrProducerClosed)
	el, _ := p.out.Drain(nil)
	for _, m := range el {
		m.fail(err)
	}

	close(p.done) // signal that producer is now closed
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

type messageWithCtx struct {
	ctx context.Context
	msg *kafka.Message
}

func (c *messageWithCtx) fail(err error) {
	c.msg.Opaque.(rx.Task[rx.Void]).Fail(err)
}
