package logx

import (
	"runtime/debug"
	"time"

	"go.uber.org/zap"
)

type Logger = *zap.Logger
type Field = zap.Field

var ErrField = zap.Error

func New() Logger {
	fields := make([]zap.Field, 0, 10)

	if info, ok := debug.ReadBuildInfo(); ok {
		fields = append(fields, zap.String("build_tools_version", info.GoVersion))
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

	logger := zap.Must(zap.NewProductionConfig().Build(zap.Fields(fields...)))
	return logger
}
