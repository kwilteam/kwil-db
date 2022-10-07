package messaging

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"kwil/x"
	cfg "kwil/x/messaging/config"
	"kwil/x/syncx"
	"kwil/x/utils"
	"os"
	"strings"
)

type producer[T any] struct {
	kp     *kafka.Producer
	topic  string
	serdes Serdes[T]
	out    syncx.Chan[*messageWithCtx]
	done   chan x.Void
}

func (p *producer[T]) Submit(ctx context.Context, message ProducerMessage[T]) error {
	m, err := p.createMessage(message)
	if err != nil {
		return err
	}

	ctx = utils.IfElse(ctx != nil, ctx, context.Background())
	if !p.out.Write(&messageWithCtx{ctx, m}) {
		return ErrProducerClosed
	}

	return nil
}

func (p *producer[T]) Close() {
	p.out.Close()
}

func (p *producer[T]) OnClosed() <-chan x.Void {
	return p.done
}

func (p *producer[T]) createMessage(message ProducerMessage[T]) (*kafka.Message, error) {
	key, payload, err := p.serdes.Serialize(message.Payload())
	if err != nil {
		return nil, err
	}

	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic},
		Key:            key,
		Value:          payload,
		Opaque:         message.GetAckNack(),
	}, nil
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
		if completeOrFail(m) {
			return
		}

		// producer is closed or some other unrecoverable error has occurred
		// CONFIRM: possible deadlock by CommitBlock() caller if close does
		// not flush out all messages to event that are pending call back
		fmt.Println("producer is closed or some other unrecoverable error has occurred")
		if done != nil {
			*done = 3
		}
	default:
		fmt.Printf("Ignored event: %s\n", e)
	}
}

func (p *producer[T]) start() {
	go p.begin_processing()
}

func (p *producer[T]) begin_processing() {
	ev := p.kp.Events()
	done := 0

	defer p.doCleanup(done, ev)

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
}

func (p *producer[T]) doCleanup(done int, ev chan kafka.Event) {
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

func loadP(config cfg.Config) (topic string, kp *kafka.Producer, err error) {
	if config == nil {
		return "", nil, fmt.Errorf("config is nil")
	}

	m := make(kafka.ConfigMap)

	settings := config.Select("broker-settings").ToStringMap()
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

	m["linger.ms"] = config.GetString("linger-ms", "50")

	if len(m) > 0 {
		fmt.Printf("using kafka consumer settings:\n")
	}

	for k, v := range m {
		if strings.Contains(k, "password") || strings.Contains(k, "secret") {
			v = "********"
		}
		fmt.Printf("\t%s=%s\n", k, v)
	}

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
	if c.msg.Opaque == nil {
		return
	}

	ackNack, ok := c.msg.Opaque.(AckNackFn)
	if ok {
		ackNack(err)
	}
}

func completeOrFail(m *kafka.Message) bool {
	if m.Opaque == nil {
		return false
	}

	ackNack, ok := m.Opaque.(AckNackFn)
	if ok {
		ackNack(m.TopicPartition.Error)
	}

	return true
}
