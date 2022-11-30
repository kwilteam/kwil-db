package sqlclient

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"ksl/sqldriver"
	"net/url"
	"sync"
)

type Client struct {
	Name string
	DB   *sql.DB
	URL  *URL

	sqldriver.Driver
	closers []io.Closer

	openDriver func(sqldriver.ExecQuerier) (sqldriver.Driver, error)
}

type TxClient struct {
	*Client
	Tx *sql.Tx
}

type URL struct {
	*url.URL
	DSN    string
	Schema string
}

// Tx returns a transactional client.
func (c *Client) Tx(ctx context.Context, opts *sql.TxOptions) (*TxClient, error) {
	if c.openDriver == nil {
		return nil, errors.New("sql/sqlclient: unexpected driver opener: <nil>")
	}
	tx, err := c.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("sql/sqlclient: starting transaction: %w", err)
	}
	drv, err := c.openDriver(tx)
	if err != nil {
		return nil, fmt.Errorf("sql/sqlclient: opening driver: %w", err)
	}
	ic := *c
	ic.Driver = drv
	return &TxClient{Client: &ic, Tx: tx}, nil
}

func (c *TxClient) Commit() error                 { return c.Tx.Commit() }
func (c *TxClient) Rollback() error               { return c.Tx.Rollback() }
func (c *Client) AddClosers(closers ...io.Closer) { c.closers = append(c.closers, closers...) }

func (c *Client) Close() (err error) {
	for _, closer := range append(c.closers, c.DB) {
		if cerr := closer.Close(); cerr != nil {
			if err != nil {
				cerr = fmt.Errorf("%v: %v", err, cerr)
			}
			err = cerr
		}
	}
	return err
}

type Opener interface {
	Open(u *url.URL) (*Client, error)
}

type OpenerFunc func(*url.URL) (*Client, error)

func (f OpenerFunc) Open(u *url.URL) (*Client, error) {
	return f(u)
}

type URLParser interface {
	ParseURL(*url.URL) *URL
}

type URLParserFunc func(*url.URL) *URL

func (f URLParserFunc) ParseURL(u *url.URL) *URL {
	return f(u)
}

type SchemaChanger interface {
	ChangeSchema(*url.URL, string) *url.URL
}

type driver struct {
	Opener
	name   string
	parser URLParser
}

var drivers sync.Map

type openOptions struct {
	schema *string
}

type OpenOption func(*openOptions) error

var ErrUnsupported = errors.New("sql/sqlclient: driver does not support changing connected schema")

func OpenProvider(provider string, s string, opts ...OpenOption) (*Client, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("sql/sqlclient: parse open url: %w", err)
	}
	return OpenURL(provider, u, opts...)
}

func Open(s string, opts ...OpenOption) (*Client, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("sql/sqlclient: parse open url: %w", err)
	}
	return OpenURL(u.Scheme, u, opts...)
}

func OpenURL(backend string, u *url.URL, opts ...OpenOption) (*Client, error) {
	cfg := &openOptions{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	v, ok := drivers.Load(backend)
	if !ok {
		return nil, fmt.Errorf("sql/sqlclient: no opener was register with name %q", u.Scheme)
	}

	if cfg.schema != nil {
		sc, ok := v.(*driver).parser.(SchemaChanger)
		if !ok {
			return nil, ErrUnsupported
		}
		u = sc.ChangeSchema(u, *cfg.schema)
	}
	client, err := v.(*driver).Open(u)
	if err != nil {
		return nil, err
	}
	if client.URL == nil {
		client.URL = v.(*driver).parser.ParseURL(u)
	}
	return client, nil
}

func OpenSchema(s string) OpenOption {
	return func(c *openOptions) error {
		c.schema = &s
		return nil
	}
}

type registerOptions struct {
	openDriver func(sqldriver.ExecQuerier) (sqldriver.Driver, error)
	parser     URLParser
	flavours   []string
}
type RegisterOption func(*registerOptions)

func RegisterFlavours(flavours ...string) RegisterOption {
	return func(opts *registerOptions) {
		opts.flavours = flavours
	}
}

func RegisterURLParser(p URLParser) RegisterOption {
	return func(opts *registerOptions) {
		opts.parser = p
	}
}

func RegisterDriverOpener(open func(sqldriver.ExecQuerier) (sqldriver.Driver, error)) RegisterOption {
	return func(opts *registerOptions) {
		opts.openDriver = open
	}
}

// Register registers a client Opener (i.e. creator) with the given name.
func Register(name string, opener Opener, opts ...RegisterOption) {
	if opener == nil {
		panic("sql/sqlclient: Register opener is nil")
	}
	opt := &registerOptions{
		parser: URLParserFunc(func(u *url.URL) *URL { return &URL{URL: u, DSN: u.String()} }),
	}
	for i := range opts {
		opts[i](opt)
	}

	if opt.openDriver != nil {
		f := opener
		opener = OpenerFunc(func(u *url.URL) (*Client, error) {
			c, err := f.Open(u)
			if err != nil {
				return nil, err
			}
			c.openDriver = opt.openDriver
			return c, err
		})
	}
	drv := &driver{opener, name, opt.parser}
	for _, f := range append(opt.flavours, name) {
		if _, ok := drivers.Load(f); ok {
			panic("sql/sqlclient: Register called twice for " + f)
		}
		drivers.Store(f, drv)
	}
}
