package log

import (
	"go.uber.org/zap"
)

// Logger is a wrapper around zap.Logger, which adds some additional fields to critical log messages.
type Logger struct {
	L *zap.Logger
}

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

func (l *Logger) clone() *Logger {
	copy := *l
	return &copy
}

func (l *Logger) Named(name string) *Logger {
	if name == "" {
		return l
	}

	_log := l.clone()
	_log.L.Named(name)
	return _log
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

type Config struct {
	Level string `mapstructure:"level"`
	// OutputPaths is a list of URLs or file paths to write logging output to.
	OutputPaths []string `mapstructure:"output_paths"`
}

func New(config Config) Logger {
	fields := make([]zap.Field, 0, 10)

	// poor man's config
	cfg := zap.NewProductionConfig()
	level, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		panic(err)
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

	logger := zap.Must(cfg.Build(zap.Fields(fields...)))
	// skip the logger wrapper
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	return Logger{L: logger}
}

// NoOp is a logger that does nothing.
// It is useful for testing, or for user packages where we want to
// have logging configurable.
func NewNoOp() Logger {
	return Logger{L: zap.NewNop()}
}
