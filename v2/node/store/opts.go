package store

import (
	"fmt"

	"p2p/log"
)

type options struct {
	logger   log.Logger
	compress bool
}

type Option func(*options)

func WithLogger(logger log.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithCompression(compress bool) Option {
	return func(o *options) {
		o.compress = compress
	}
}

// badgerLogger implements the badger.Logger interface.
type badgerLogger struct {
	log log.Logger
}

func (b *badgerLogger) Debugf(p0 string, p1 ...any) {
	b.log.Debug(fmt.Sprintf(p0, p1...))
}

func (b *badgerLogger) Errorf(p0 string, p1 ...any) {
	b.log.Error(fmt.Sprintf(p0, p1...))
}

func (b *badgerLogger) Infof(p0 string, p1 ...any) {
	b.log.Info(fmt.Sprintf(p0, p1...))
}

func (b *badgerLogger) Warningf(p0 string, p1 ...any) {
	b.log.Warn(fmt.Sprintf(p0, p1...))
}
