package log

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger, which adds some additional fields to critical log messages.
type Logger struct {
	L *zap.Logger
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
	return &Logger{l.L.Named(name)}
}

func (l *Logger) With(fields ...Field) *Logger {
	return &Logger{l.L.With(fields...)}
}

func (l *Logger) WithOptions(opts ...zap.Option) *Logger {
	return &Logger{l.L.WithOptions(opts...)}
}

func (l *Logger) Sync() error {
	return l.L.Sync()
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

type Config struct {
	Level string
	// OutputPaths is a list of URLs or file paths to write logging output to.
	OutputPaths []string
	// Format is either EncodingJSON or EncodingConsole.
	Format string
	// EncodeTime indicates how to encode the time. The default is
	// TimeEncodingEpochFloat.
	EncodeTime string
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
	// poor man's config
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
	switch enc := config.Format; enc {
	case FormatPlain: // "plain" => "console"
		cfg.Encoding = "console"
	case FormatJSON, "": // also the default
		cfg.Encoding = "json"
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

	// fields := make([]zap.Field, 0, 10) with cfg.Build(zap.Fields(fields...))
	logger := zap.Must(cfg.Build())
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	return Logger{L: logger}, nil
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
	s.S.Debugf(msg, fields...)
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
