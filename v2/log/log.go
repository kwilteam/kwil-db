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

type KVLogger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Log(level Level, msg string, args ...any)
}

type Loggerf interface {
	Debugf(msg string, args ...any)
	Infof(msg string, args ...any)
	Warnf(msg string, args ...any)
	Errorf(msg string, args ...any)
	Logf(level Level, msg string, args ...any)
}

type Loggerln interface {
	Debugln(a ...any)
	Infoln(a ...any)
	Warnln(a ...any)
	Errorln(a ...any)
	Logln(level Level, a ...any)
}

type LoggerMaker interface {
	// New creates a child logger using the same backend and options as the
	// current logger, but with the specified name and level.
	New(name string) Logger
	NewWithLevel(lvl Level, name string) Logger
}

type Logger interface {
	KVLogger // (msg string, args ...any) where args are key-value pairs
	Loggerf  // (msg string, args ...any) where args are printf like arguments
	Loggerln // (a ...any) in the manner of println

	LoggerMaker
}

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the log level. Use [ParseLevel]
// to go from a string to a Level.
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

func (lvl Level) MarshalText() (text []byte, err error) {
	switch lvl {
	case LevelDebug:
		return []byte("debug"), nil
	case LevelInfo:
		return []byte("info"), nil
	case LevelWarn:
		return []byte("warn"), nil
	case LevelError:
		return []byte("error"), nil
	default:
		return nil, errors.New("unknown log level: " + lvl.String())
	}
}

func (lvl *Level) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "debug":
		*lvl = LevelDebug
	case "info":
		*lvl = LevelInfo
	case "warn":
		*lvl = LevelWarn
	case "error":
		*lvl = LevelError
	default:
		return errors.New("unknown log level: " + string(text))
	}
	return nil
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

// ParseLevel parses a string into a log level. Use [Level.String] to go from a
// Level to a string.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
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
		if i == 2 {
			sp = " "
		}
		// if odd, then we will just print the last value
		if i+1 >= len(args) {
			fmt.Fprintf(&msg, "%s%v", sp, args[i])
			break
		}
		key, val := args[i], args[i+1]
		if v, ok := val.([]byte); ok {
			val = hex.EncodeToString(v)
		}
		fmt.Fprintf(&msg, "%s%s=%v", sp, key, val)
	}
	msg.WriteString("}")
	return msg.String()
}

var _ KVLogger = (*plainLogger)(nil)

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

var _ Loggerln = (*plainLogger)(nil)

func (l *plainLogger) Debugln(a ...any) {
	l.log.Debug(a...)
}

func (l *plainLogger) Infoln(a ...any) {
	l.log.Info(a...)
}

func (l *plainLogger) Warnln(a ...any) {
	l.log.Warn(a...)
}

func (l *plainLogger) Errorln(a ...any) {
	l.log.Error(a...)
}

func (l *plainLogger) Logln(level Level, a ...any) {
	switch level {
	case LevelDebug:
		l.Debugln(a...)
	case LevelInfo:
		l.Infoln(a...)
	case LevelWarn:
		l.Warnln(a...)
	case LevelError:
		l.Errorln(a...)
	}
}

var _ Loggerf = (*plainLogger)(nil)

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

var _ LoggerMaker = (*plainLogger)(nil)

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

// slog has reserved keys ("time", "level", "msg", and "source"), and it is
// almost always unintended to override these values. Further, slog will
// actually panic in cases where the value type is not as expected e.g. not a
// time.Time for the "time" key! This helper ensures the application corrupt the
// logs or crash in the event of a buggy log line.
func sanitizeArgs(args []any) []any {
	sanitized := make([]any, 0, len(args))
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue // Skip if key is not a string
		}

		// Rename reserved keys to avoid conflict with slog's internal fields.
		// If you see these underscore-suffixed keys in the logs, fix your log.
		switch key {
		case slog.TimeKey:
			key = "time_"
		case slog.LevelKey:
			key = "level_"
		case slog.MessageKey: // "msg"
			key = "msg_"
		case slog.SourceKey:
			key = "source_"
		}

		sanitized = append(sanitized, key)
		if i+1 < len(args) { // don't panic if args len was odd
			sanitized = append(sanitized, args[i+1])
		}
	}
	return sanitized
}

var _ KVLogger = (*logger)(nil)

func (l *logger) Debug(msg string, args ...any) {
	l.log.Debug(msg, sanitizeArgs(args)...)
}
func (l *logger) Info(msg string, args ...any) {
	l.log.Info(msg, sanitizeArgs(args)...)
}
func (l *logger) Warn(msg string, args ...any) {
	l.log.Warn(msg, sanitizeArgs(args)...)
}
func (l *logger) Error(msg string, args ...any) {
	l.log.Error(msg, sanitizeArgs(args)...)
}
func (l *logger) Log(level Level, msg string, args ...any) {
	l.log.Log(context.Background(), levelToSlog(level), msg, sanitizeArgs(args)...)
}

var _ Loggerf = (*logger)(nil)

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

var _ Loggerln = (*logger)(nil)

func (l *logger) Debugln(a ...any) {
	l.Debug(fmt.Sprintln(a...))
}
func (l *logger) Infoln(a ...any) {
	l.Info(fmt.Sprintln(a...))
}
func (l *logger) Warnln(a ...any) {
	l.Warn(fmt.Sprintln(a...))
}
func (l *logger) Errorln(a ...any) {
	l.Error(fmt.Sprintln(a...))
}
func (l *logger) Logln(level Level, a ...any) {
	l.Log(level, fmt.Sprintln(a...))
}

var _ LoggerMaker = (*logger)(nil)

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

func (l *discardLogger) Debugln(a ...any)            {}
func (l *discardLogger) Infoln(a ...any)             {}
func (l *discardLogger) Warnln(a ...any)             {}
func (l *discardLogger) Errorln(a ...any)            {}
func (l *discardLogger) Logln(level Level, a ...any) {}

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
