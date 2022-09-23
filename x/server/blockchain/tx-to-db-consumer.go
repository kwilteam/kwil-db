package blockchain

import (
	"kwil/x/common/host"
	cfg "kwil/x/common/host/config"
	types "kwil/x/common/host/service"
)

var txToDbConsumerIdentity types.ServiceIdentity

const CHAIN_CONSUMER_SERVICE_NAME string = "tx-to-db-consumer-service"

func init() {
	identity, err := host.RegisterService(CHAIN_CONSUMER_SERVICE_NAME, func() types.Service {
		return &txToDbConsumer{
			quit: make(chan struct{}),
		}
	})
	if err != nil {
		panic(err)
	}
	txToDbConsumerIdentity = identity
}

// ToD: This consumer will translate the chain tx into a SQL command
// and send it to pgSql for execution
type txToDbConsumer struct {
	types.BackgroundService
	quit chan struct{}
}

func (c *txToDbConsumer) Identity() types.ServiceIdentity {
	return txToDbConsumerIdentity
}

func (c *txToDbConsumer) Configure(config cfg.Config) error {
	panic("not implemented")
}

func (c *txToDbConsumer) Initialize(ctx types.ServiceContext) error {
	panic("not implemented")
}

func (c *txToDbConsumer) Start(ctx types.ServiceContext) error {
	panic("not implemented")
}

func (c *txToDbConsumer) Shutdown() {
	c.quit <- struct{}{}

	q := c.quit
	c.quit = nil
	close(q)
}

func (c *txToDbConsumer) IsRunning() bool {
	return c.quit != nil
}

func (c *txToDbConsumer) AwaitShutdown() {
	<-c.quit
}
