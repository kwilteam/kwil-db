package log

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	sublog "github.com/decred/slog"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (lvl Level) String() string {
	switch lvl {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

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

func ParseLevel(s string) (Level, error) {
	switch s {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return 0, errors.New("unknown log level: " + s)
	}
}

// plainLogger is a plain text logger (not structured)
type plainLogger struct {
	be  *sublog.Backend
	log sublog.Logger
}

// args ...any, this must become arg[0]=arg[1] etc in the printed message
func formatArgs(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	var sp string
	var msg strings.Builder
	msg.WriteString(" {")
	// args are pairs of key-values, so we will print them in pairs after the message.
	for i := 0; i < len(args); i += 2 {
		// if odd, then we will just print the last value
		if i+1 >= len(args) {
			fmt.Fprintf(&msg, " %v", args[i])
			break
		}
		key, val := args[i], args[i+1]
		if v, ok := val.([]byte); ok {
			val = hex.EncodeToString(v)
		}
		fmt.Fprintf(&msg, "%s%s=%v", sp, key, val)
		if i == 0 {
			sp = " "
		}
	}
	msg.WriteString("}")
	return msg.String()
}

func (l *plainLogger) Debug(msg string, args ...any) {
	// args are pairs of key-values, so we will print them in pairs after the message.
	msg += formatArgs(args...)
	l.log.Debugf(msg)
}
func (l *plainLogger) Info(msg string, args ...any) {
	msg += formatArgs(args...)
	l.log.Infof(msg)
}
func (l *plainLogger) Warn(msg string, args ...any) {
	msg += formatArgs(args...)
	l.log.Warnf(msg)
}
func (l *plainLogger) Error(msg string, args ...any) {
	msg += formatArgs(args...)
	l.log.Errorf(msg)
}
func (l *plainLogger) Log(level Level, msg string, args ...any) {
	switch level {
	case LevelDebug:
		l.Debug(msg, args...)
	case LevelInfo:
		l.Info(msg, args...)
	case LevelWarn:
		l.Warn(msg, args...)
	case LevelError:
		l.Error(msg, args...)
	}
}
func (l *plainLogger) Debugf(msg string, args ...any) {
	l.log.Debugf(msg, args...)
}
func (l *plainLogger) Infof(msg string, args ...any) {
	l.log.Infof(msg, args...)
}
func (l *plainLogger) Warnf(msg string, args ...any) {
	l.log.Warnf(msg, args...)
}

func (l *plainLogger) Errorf(msg string, args ...any) {
	l.log.Errorf(msg, args...)
}
func (l *plainLogger) Logf(level Level, msg string, args ...any) {
	switch level {
	case LevelDebug:
		l.Debugf(msg, args...)
	case LevelInfo:
		l.Infof(msg, args...)
	case LevelWarn:
		l.Warnf(msg, args...)
	case LevelError:
		l.Errorf(msg, args...)
	}
}

func (l *plainLogger) NewWithLevel(lvl Level, name string) Logger {
	logger := l.be.Logger(name)
	logger.SetLevel(levelToSublog(lvl))
	return &plainLogger{
		be:  l.be,
		log: logger,
	}
}
func (l *plainLogger) New(name string) Logger {
	logger := l.be.Logger(name)
	return &plainLogger{
		be:  l.be,
		log: logger,
	}
}

func levelToSublog(l Level) sublog.Level {
	// convert Level to sublog level
	switch l {
	case LevelDebug:
		return sublog.LevelDebug
	case LevelInfo:
		return sublog.LevelInfo
	case LevelWarn:
		return sublog.LevelWarn
	case LevelError:
		return sublog.LevelError
	default:
		return sublog.LevelInfo
	}
}

// logger is a structured logger
type logger struct {
	hOpts slog.HandlerOptions
	opts  options
	log   *slog.Logger
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
func (l *logger) NewWithLevel(lvl Level, name string) Logger {
	opts := l.opts
	opts.name = name
	opts.level = lvl
	return newLogger(&opts)
}
func (l *logger) New(name string) Logger {
	opts := l.opts
	opts.name = name
	return newLogger(&opts)
}

func (l *logger) WithGroup(group string) Logger {
	return &logger{
		log: l.log.WithGroup(group),
	}
}

type LoggerMaker interface {
	// New creates a child logger using the same backend and options as the
	// current logger, but with the specified name and level.
	New(name string) Logger
	NewWithLevel(lvl Level, name string) Logger
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

	LoggerMaker

	// With
	// WithGroup(group string) Logger
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

//	func (l *discardLogger) WithGroup(group string) Logger {
//		return &discardLogger{}
//	}
func (l *discardLogger) NewWithLevel(lvl Level, name string) Logger {
	return &discardLogger{}
}
func (l *discardLogger) New(name string) Logger {
	return &discardLogger{}
}
func (l *discardLogger) Debugf(msg string, args ...any)            {}
func (l *discardLogger) Infof(msg string, args ...any)             {}
func (l *discardLogger) Warnf(msg string, args ...any)             {}
func (l *discardLogger) Errorf(msg string, args ...any)            {}
func (l *discardLogger) Logf(level Level, msg string, args ...any) {}

func New(opts ...Option) Logger {
	options := &options{
		name:      "",
		level:     LevelInfo,
		addSource: false,
		writer:    os.Stdout,
		format:    "text",
	}

	for _, opt := range opts {
		opt(options)
	}

	return newLogger(options)
}

func newLogger(options *options) Logger {
	if options.writer == nil {
		options.writer = os.Stdout
	}

	if options.format == FormatUnstructured {
		be := sublog.NewBackend(options.writer)
		logger := be.Logger(options.name)
		logger.SetLevel(levelToSublog(options.level))
		return &plainLogger{
			be:  be,
			log: logger,
		}
	}

	if options.format == "" {
		options.format = "text"
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
	switch options.format {
	case FormatJSON:
		handler = slog.NewJSONHandler(options.writer, handlerOpts)
	case FormatText:
		handler = slog.NewTextHandler(options.writer, handlerOpts)
	default:
		panic(fmt.Sprintf("unknown logging format: %s", options.format))
	}

	// if options.name != "" { // name => group for stdlib log/slog
	// 	handler = handler.WithGroup(options.name)
	// }
	if options.name != "" {
		handler = handler.WithAttrs([]slog.Attr{
			slog.String("system", options.name),
		})
	}

	return &logger{
		hOpts: *handlerOpts,
		opts:  *options,
		log:   slog.New(handler),
	}
}
