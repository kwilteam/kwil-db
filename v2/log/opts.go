package log

import "io"

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type Option func(*options)

type options struct {
	level     Level
	addSource bool
	writer    io.Writer
	format    Format
	group     string // slog group for WithGroup, like a namespace
}

func WithGroup(group string) Option {
	return func(o *options) {
		o.group = group
	}
}

func WithLevel(level Level) Option {
	return func(o *options) {
		o.level = level
	}
}

func WithSource(enabled bool) Option {
	return func(o *options) {
		o.addSource = enabled
	}
}

func WithWriter(w io.Writer) Option {
	return func(o *options) {
		o.writer = w
	}
}

func WithFormat(format Format) Option {
	return func(o *options) {
		o.format = format
	}
}
