package log

import (
	"go.uber.org/zap"
	"runtime/debug"
	"time"
)

type Field = zap.Field

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
