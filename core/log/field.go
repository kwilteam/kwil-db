package log

import (
	"runtime/debug"
	"time"

	"go.uber.org/zap"
)

type Field = zap.Field

func String[T ~string](key string, val T) Field { // work with types that are an underlying string
	return zap.String(key, string(val))
}

func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

func Float[T ~float32 | ~float64](key string, val T) Field {
	return zap.Float64(key, float64(val))
}

func Int[T ~int | ~int64 | ~int32 | ~int16 | ~int8](key string, val T) Field {
	return zap.Int64(key, int64(val))
}

func Uint[T ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8](key string, val T) Field {
	return zap.Uint64(key, uint64(val))
}

func Any(key string, val any) Field {
	return zap.Any(key, val)
}

func Error(err error) Field {
	return zap.Error(err)
}

func genDetailedField() []zap.Field {
	var fields []zap.Field
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
	return fields
}

var detailedFields []Field = genDetailedField()
