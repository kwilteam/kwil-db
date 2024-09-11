package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jrick/logrotate/rotator"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger, which adds some additional fields to critical log messages.
type Logger struct {
	L     *zap.Logger
	close func() error
}

func (l *Logger) Level() Level {
	return l.L.Level()
}

// non-structured logging

func (l *Logger) Debugf(msg string, args ...any) {
	l.L.Debug(fmt.Sprintf(msg, args...))
}

func (l *Logger) Infof(msg string, args ...any) {
	l.L.Info(fmt.Sprintf(msg, args...))
}

func (l *Logger) Warnf(msg string, args ...any) {
	l.L.Warn(fmt.Sprintf(msg, args...))
}

func (l *Logger) Errorf(msg string, args ...any) {
	l.L.Error(fmt.Sprintf(msg, args...))
}

func (l *Logger) Logf(level Level, msg string, args ...any) {
	l.L.Log(level, fmt.Sprintf(msg, args...))
}

// structured logging

func (l *Logger) Debug(msg string, fields ...Field) {
	l.L.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.L.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.L.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	fields = append(fields, detailedFields...)
	l.L.Error(msg, fields...)
}

func (l *Logger) DPanic(msg string, fields ...Field) {
	fields = append(fields, detailedFields...)
	l.L.DPanic(msg, fields...)
}

func (l *Logger) Panic(msg string, fields ...Field) {
	fields = append(fields, detailedFields...)
	l.L.Panic(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	fields = append(fields, detailedFields...)
	l.L.Fatal(msg, fields...)
}

func (l *Logger) Log(level Level, msg string, fields ...Field) {
	if level >= ErrorLevel {
		fields = append(fields, detailedFields...)
	}
	l.L.Log(level, msg, fields...)
}

func (l *Logger) Named(name string) *Logger {
	if name == "" {
		return l
	}
	return &Logger{L: l.L.Named(name), close: l.close}
}

func (l *Logger) With(fields ...Field) *Logger {
	return &Logger{L: l.L.With(fields...), close: l.close}
}

func (l *Logger) WithOptions(opts ...zap.Option) *Logger {
	return &Logger{L: l.L.WithOptions(opts...), close: l.close}
}

// IncreasedLevel creates a logger clone with a higher log level threshold,
// which is ignored if it is lower than the parent logger's level.
func (l *Logger) IncreasedLevel(lvl Level) *Logger {
	return l.WithOptions(zap.IncreaseLevel(zap.NewAtomicLevelAt(lvl)))
}

func (l *Logger) Sync() error {
	return l.L.Sync()
}

var _ io.Closer = (*Logger)(nil)
var _ io.Closer = Logger{}

func (l Logger) Close() error {
	if l.close == nil {
		return nil
	}
	return l.close()
}

func (l Logger) Sugar() SugaredLogger {
	return SugaredLogger{S: l.L.Sugar()}
}

const (
	FormatJSON  = "json"
	FormatPlain = "plain"
)

const (
	TimeEncodingRFC3339Milli = "rfc3339milli"
	TimeEncodingEpochMilli   = "epochmilli"
	TimeEncodingEpochFloat   = "epochfloat"
)

const (
	// defaults for log data retained:
	// - 6 GB total uncompressed in 100 (minus) gzipped files
	// - up to 60 MB in current uncompressed log
	maxLogRolls  = 100
	maxLogSizeKB = 60_000
)

type Config struct {
	Level string
	// OutputPaths is a list of URLs or file paths to write logging output to.
	OutputPaths []string
	// Format is either EncodingJSON or EncodingConsole.
	Format string
	// EncodeTime indicates how to encode the time. The default is
	// TimeEncodingEpochFloat.
	EncodeTime string

	MaxLogRolls int

	MaxLogSizeKB int64
}

func rfc3339MilliTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.999Z07:00"))
}

func epochMillisTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.UnixMilli())
}

func New(config Config) Logger {
	logger, err := NewChecked(config)
	if err != nil {
		panic(err.Error())
	}
	return logger
}

func NewChecked(config Config) (Logger, error) {
	cfg := zap.NewProductionConfig()
	level, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		return Logger{}, err
	}

	switch config.EncodeTime {
	case TimeEncodingRFC3339Milli:
		cfg.EncoderConfig.EncodeTime = rfc3339MilliTimeEncoder
	case TimeEncodingEpochMilli:
		cfg.EncoderConfig.EncodeTime = epochMillisTimeEncoder
	case TimeEncodingEpochFloat: // this is the default
		fallthrough
	default:
		// no-op, the default from NewProductionConfig in zap 1.25, but set it
		// to ensure stability.
		cfg.EncoderConfig.EncodeTime = zapcore.EpochTimeEncoder
	}

	// Translate from our formats to Zap's
	var zapEncMaker func(cfg zapcore.EncoderConfig) zapcore.Encoder
	switch enc := config.Format; enc {
	case FormatPlain: // "plain" => "console"
		cfg.Encoding = "console"
		zapEncMaker = zapcore.NewConsoleEncoder
	case FormatJSON, "": // also the default
		cfg.Encoding = "json"
		zapEncMaker = zapcore.NewJSONEncoder
	default:
		return Logger{}, fmt.Errorf("invalid log format %q", enc)
	}

	if config.Level != "" {
		cfg.Level = level
	} else {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	if len(config.OutputPaths) > 0 {
		cfg.OutputPaths = config.OutputPaths
	} else {
		cfg.OutputPaths = []string{"stdout"}
	}

	if config.MaxLogRolls == 0 {
		config.MaxLogRolls = maxLogRolls
	}
	if config.MaxLogSizeKB == 0 {
		config.MaxLogSizeKB = maxLogSizeKB
	}

	var wss wsMultiWrapper
	for _, w := range cfg.OutputPaths {
		switch w {
		case "stdout":
			wss.tees = append(wss.tees, &noCloser{os.Stdout})
		case "stderr":
			wss.tees = append(wss.tees, &noCloser{os.Stderr})
		default: // log file
			rotator, err := rotator.New(w, config.MaxLogSizeKB,
				false, config.MaxLogRolls)
			if err != nil {
				return Logger{}, err
			}
			wss.tees = append(wss.tees, rotator)
		}
	}

	var writeSyncer zapcore.WriteSyncer = &wss

	enc := zapEncMaker(cfg.EncoderConfig)
	logCore := zapcore.NewCore(enc, writeSyncer, cfg.Level)
	logger := zap.New(logCore)

	logger = logger.WithOptions(zap.AddCallerSkip(1))
	return Logger{L: logger, close: wss.Close}, nil
}

// noCloser embeds all the methods of a zapcore.WriteSyncer but hides any Close
// method. This is used for os.Stdout and os.Stderr, which have a Close method,
// but which should not be closed as per os docs.
type noCloser struct{ zapcore.WriteSyncer }

type wsMultiWrapper struct {
	tees []io.Writer
}

var _ io.Writer = (*wsMultiWrapper)(nil)

func (wss *wsMultiWrapper) Write(b []byte) (int, error) {
	var err error
	var n int
	for _, w := range wss.tees {
		ni, erri := w.Write(b)
		err = errors.Join(err, erri)
		n = max(ni, n)
	}
	return n, err
}

var _ zapcore.WriteSyncer = (*wsMultiWrapper)(nil)

func (wss *wsMultiWrapper) Sync() error {
	var err error
	for _, w := range wss.tees {
		if s, ok := w.(zapcore.WriteSyncer); ok {
			err = errors.Join(err, s.Sync())
		}
	}
	return err
}

var _ io.Closer = (*wsMultiWrapper)(nil)

func (wss *wsMultiWrapper) Close() error {
	var err error
	for _, w := range wss.tees {
		if s, ok := w.(io.Closer); ok {
			err = errors.Join(err, s.Close())
		}
	}
	return err
}

type Level = zapcore.Level

const (
	DebugLevel  Level = zap.DebugLevel
	InfoLevel         = zap.InfoLevel
	WarnLevel         = zap.WarnLevel
	ErrorLevel        = zap.ErrorLevel
	DPanicLevel       = zap.DPanicLevel
	PanicLevel        = zap.PanicLevel
	FatalLevel        = zap.FatalLevel
)

// ParseLevel parses a log level string, which is useful for reading level
// settings from a config file or other text source.
func ParseLevel(lvl string) (Level, error) {
	l, err := zapcore.ParseLevel(lvl)
	if err != nil {
		return zapcore.InvalidLevel, err
	}
	return l, nil
}

func NewStdOut(level Level) Logger {
	return New(Config{
		Level:       level.String(),
		OutputPaths: []string{"stdout"},
	})
}

// NoOp is a logger that does nothing.
// It is useful for testing, or for user packages where we want to
// have logging configurable.
func NewNoOp() Logger {
	return Logger{L: zap.NewNop()}
}

type SugaredLogger struct {
	S *zap.SugaredLogger
}

func (s *SugaredLogger) Level() Level {
	return s.S.Level()
}

func (s *SugaredLogger) Debug(msg string, fields ...any) {
	s.S.Debugw(msg, fields...)
}

func (s *SugaredLogger) Info(msg string, fields ...any) {
	s.S.Infow(msg, fields...)
}

func (s *SugaredLogger) Warn(msg string, fields ...any) {
	s.S.Warnw(msg, fields...)
}

func (s *SugaredLogger) Error(msg string, fields ...any) {
	s.S.Errorw(msg, fields...)
}

func (s *SugaredLogger) Named(name string) *SugaredLogger {
	return &SugaredLogger{S: s.S.Named(name)}
}

func (s *SugaredLogger) Sync() error {
	return s.S.Sync()
}
