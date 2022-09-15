package blockchain

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/kwilteam/kwil-db/internal/common/host"
	cfg "github.com/kwilteam/kwil-db/internal/common/host/config"
	types "github.com/kwilteam/kwil-db/internal/common/host/service"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var chainProducerIdentity types.ServiceIdentity

const CHAIN_PRODUCER_SERVICE_NAME string = "tx-to-topic-producer-service"

func init() {
	identity, err := host.RegisterService(CHAIN_PRODUCER_SERVICE_NAME, func() types.Service {
		return &chainProducerImpl{
			quit: make(chan struct{}),
		}
	})
	if err != nil {
		panic(err)
	}
	chainProducerIdentity = identity
}

type chainProducerImpl struct {
	types.ClosableService
	producer *kafka.Producer
	topic    string
	quit     chan struct{}
}

func (c *chainProducerImpl) Identity() types.ServiceIdentity {
	return chainProducerIdentity
}

func (c *chainProducerImpl) CommitBlock(cb *ChainTxCallback) {
	err := c.producer.Produce(c.createMessage(cb), nil)
	if err != nil {
		cb.Error(err)
	}
}

func (c *chainProducerImpl) Configure(config cfg.Config) error {
	return c.configure_internal(config)
}

func (c *chainProducerImpl) Initialize(ctx types.ServiceContext) error {
	s, err := ctx.GetServiceById(chainIdentity.Id())
	if err != nil {
		return err
	}

	c.begin_event_processing()

	s.(Chain).SetHandler(c)

	return nil
}

func (c *chainProducerImpl) Close() {
	c.producer.Close()
	c.quit <- struct{}{}

	q := c.quit
	c.quit = nil
	close(q)
}

func (c *chainProducerImpl) createMessage(cb *ChainTxCallback) *kafka.Message {
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &c.topic},
		Key:            cb.ctx.tx().group,
		Value:          cb.ctx.tx().data,
		Opaque:         cb,
	}
}

func (c *chainProducerImpl) begin_event_processing() {
	go func() {
		ev := c.producer.Events()
		for {
			select {
			case <-c.quit:
				return

			case e, ok := <-ev:
				if !ok {
					return
				}

				switch m := e.(type) {
				case *kafka.Message:
					if m.Opaque == nil {
						// Producer is closed or some other unrecoverable error has occured
						// CONFIRM: possible deadlock by CommitBlock() caller if close does not flush out all messages to event that are pending call back
						return //TODO: log error
					}

					cb := m.Opaque.(*ChainTxCallback)
					if m.TopicPartition.Error != nil {
						cb.Error(m.TopicPartition.Error)
					} else {
						cb.Success()
					}
				}
			}
			if c.quit == nil {
				return
			}
		}
	}()
}

func (c *chainProducerImpl) configure_internal(config cfg.Config) error {
	m := make(kafka.ConfigMap)

	settings := config.Select("cluster-settings").ToMap()
	for k, v := range settings {
		m[k] = kafka.ConfigValue(v)
	}

	c.topic = config.String("topic")
	if c.topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}

	if _, ok := settings["client.id"]; !ok {
		h, _ := os.Hostname()
		m["client.id"] = h + "_" + uuid.New().String()
	}

	m["linger.ms"] = config.Int32("linger-ms", 50)

	p, err := kafka.NewProducer(&m)
	if err != nil {
		return fmt.Errorf("failed to create producer: %s", err)
	}

	c.producer = p

	return nil
}
