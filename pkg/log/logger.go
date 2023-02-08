package log

import (
	"runtime/debug"
	"time"

	"go.uber.org/zap"
)

type SugaredLogger = *zap.SugaredLogger
type Logger = *zap.Logger
type Field = zap.Field

type Config struct {
	Level string `mapstructure:"level"`
	// OutputPaths is a list of URLs or file paths to write logging output to.
	OutputPaths []string `mapstructure:"output_paths"`
}

func New(config Config) Logger {
	// @yaiba TODO: make those only for error level?
	fields := make([]zap.Field, 0, 10)
	if info, ok := debug.ReadBuildInfo(); ok {
		fields = append(fields, zap.String("goversion", info.GoVersion))
		if info.Main.Version != "" {
			fields = append(fields, zap.String("mod_version", info.Main.Version))
		}

		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				fields = append(fields, zap.String("revision", kv.Value))
			case "vcs.time":
				if t, err := time.Parse(time.RFC3339, kv.Value); err == nil {
					fields = append(fields, zap.Time("build_time", t))
				}
			case "vcs.modified":
				if kv.Value == "true" {
					fields = append(fields, zap.Bool("dirty", true))
				}
			}
		}
	}

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
	return logger
}
