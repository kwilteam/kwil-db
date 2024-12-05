package node

import (
	"context"
	"crypto/tls"
	"errors"
	"slices"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node/pg"
)

// dbOpener opens a sessioned database connection.  Note that in this function the
// dbName is not a Kwil dataset, but a database that can contain multiple
// datasets in different postgresql "schema".
type dbOpener func(ctx context.Context, dbName string, maxConns uint32) (*pg.DB, error)

func newDBOpener(host, port, user, pass string) dbOpener {
	return func(ctx context.Context, dbName string, maxConns uint32) (*pg.DB, error) {
		cfg := &pg.DBConfig{
			PoolConfig: pg.PoolConfig{
				ConnConfig: pg.ConnConfig{
					Host:   host,
					Port:   port,
					User:   user,
					Pass:   pass,
					DBName: dbName,
				},
				MaxConns: maxConns,
			},
		}
		return pg.NewDB(ctx, cfg)
	}
}

// poolOpener opens a basic database connection pool.
type poolOpener func(ctx context.Context, dbName string, maxConns uint32) (*pg.Pool, error)

func newPoolBOpener(host, port, user, pass string) poolOpener {
	return func(ctx context.Context, dbName string, maxConns uint32) (*pg.Pool, error) {
		cfg := &pg.PoolConfig{
			ConnConfig: pg.ConnConfig{
				Host:   host,
				Port:   port,
				User:   user,
				Pass:   pass,
				DBName: dbName,
			},
			MaxConns: maxConns,
		}
		return pg.NewPool(ctx, cfg)
	}
}

type coreDependencies struct {
	ctx        context.Context
	rootDir    string
	cfg        *config.Config
	genesisCfg *config.GenesisConfig
	privKey    crypto.PrivateKey

	adminKey *tls.Certificate
	// autogen  bool

	logger     log.Logger
	dbOpener   dbOpener
	poolOpener poolOpener
}

// newService returns a common.Service with the given logger name
func (c *coreDependencies) newService(loggerName string) *common.Service {
	return &common.Service{
		Logger:        c.logger.New(loggerName),
		GenesisConfig: c.genesisCfg,
		LocalConfig:   c.cfg,
		Identity:      c.privKey.Public().Bytes(),
	}
}

// closeFuncs holds a list of closers
// it is used to close all resources on shutdown
type closeFuncs struct {
	closers []func() error
	logger  log.Logger
}

func (c *closeFuncs) addCloser(f func() error, msg string) {
	// push to top of stack
	c.closers = slices.Insert(c.closers, 0, func() error {
		c.logger.Info(msg)
		return f()
	})
}

// closeAll closes all closers
func (c *closeFuncs) closeAll() error {
	var err error
	for _, closer := range c.closers {
		err = errors.Join(closer())
	}

	return err
}

// panicErr is the type given to panic from failBuild so that the wrapped error
// may be type-inspected.
type panicErr struct {
	err error
	msg string
}

func (pe panicErr) String() string {
	return pe.msg
}

func (pe panicErr) Error() string { // error interface
	return pe.msg
}

func (pe panicErr) Unwrap() error {
	return pe.err
}
