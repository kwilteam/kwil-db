package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func levelToSlog(l Level) slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type logger struct {
	log *slog.Logger
}

func (l *logger) Debug(msg string, args ...any) {
	l.log.Debug(msg, args...)
}
func (l *logger) Info(msg string, args ...any) {
	l.log.Info(msg, args...)
}
func (l *logger) Warn(msg string, args ...any) {
	l.log.Warn(msg, args...)
}
func (l *logger) Error(msg string, args ...any) {
	l.log.Error(msg, args...)
}
func (l *logger) Log(level Level, msg string, args ...any) {
	l.log.Log(context.Background(), levelToSlog(level), msg, args...)
}
func (l *logger) Debugf(msg string, args ...any) {
	l.log.Debug(fmt.Sprintf(msg, args...))
}
func (l *logger) Infof(msg string, args ...any) {
	l.log.Info(fmt.Sprintf(msg, args...))
}
func (l *logger) Warnf(msg string, args ...any) {
	l.log.Warn(fmt.Sprintf(msg, args...))
}
func (l *logger) Errorf(msg string, args ...any) {
	l.log.Error(fmt.Sprintf(msg, args...))
}
func (l *logger) Logf(level Level, msg string, args ...any) {
	l.log.Log(context.Background(), levelToSlog(level), fmt.Sprintf(msg, args...))
}
func (l *logger) WithGroup(group string) Logger {
	return &logger{
		log: l.log.WithGroup(group),
	}
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Log(level Level, msg string, args ...any)
	Debugf(msg string, args ...any)
	Infof(msg string, args ...any)
	Warnf(msg string, args ...any)
	Errorf(msg string, args ...any)
	Logf(level Level, msg string, args ...any)
	// With
	WithGroup(group string) Logger
}

var defaultHandlerOpts = &slog.HandlerOptions{
	AddSource: true,
	Level:     slog.LevelInfo,
}

var defaultHandler = slog.NewTextHandler(os.Stdout, defaultHandlerOpts)

var defaultSLogger = slog.New(defaultHandler)

func NewStdoutLogger() Logger {
	return New(WithWriter(os.Stdout))
}

var DiscardLogger Logger = &discardLogger{} // New(WithWriter(io.Discard))

type discardLogger struct{}

func (l *discardLogger) Debug(msg string, args ...any)            {}
func (l *discardLogger) Info(msg string, args ...any)             {}
func (l *discardLogger) Warn(msg string, args ...any)             {}
func (l *discardLogger) Error(msg string, args ...any)            {}
func (l *discardLogger) Log(level Level, msg string, args ...any) {}
func (l *discardLogger) WithGroup(group string) Logger {
	return &discardLogger{}
}
func (l *discardLogger) Debugf(msg string, args ...any)            {}
func (l *discardLogger) Infof(msg string, args ...any)             {}
func (l *discardLogger) Warnf(msg string, args ...any)             {}
func (l *discardLogger) Errorf(msg string, args ...any)            {}
func (l *discardLogger) Logf(level Level, msg string, args ...any) {}

func New(opts ...Option) Logger {
	options := &options{
		level:     LevelInfo,
		addSource: false,
		writer:    os.Stdout,
		format:    "text",
		group:     "",
	}

	for _, opt := range opts {
		opt(options)
	}

	handlerOpts := &slog.HandlerOptions{
		AddSource: options.addSource,
		Level:     levelToSlog(options.level),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey { // reformat the "time" attribute
				t := a.Value.Time() // time.Now().UTC()
				a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05.000"))
			}
			return a
		},
	}

	var handler slog.Handler
	if options.format == "json" {
		handler = slog.NewJSONHandler(options.writer, handlerOpts)
	} else {
		handler = slog.NewTextHandler(options.writer, handlerOpts)
	}
	if options.group != "" {
		handler = handler.WithGroup(options.group)
	}

	return &logger{slog.New(handler)}
}
