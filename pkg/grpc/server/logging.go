package server

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
)

// InterceptorLogger adapts zap logger to interceptor logger.
// This code is copied from go-grpc-middleware/interceptors/logging/examples/zap/example_test.go
func InterceptorLogger(l *log.Logger) logging.Logger {

	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		var f []zap.Field
		i := logging.Fields(fields).Iterator()
		for i.Next() {
			k, v := i.At()
			f = append(f, zap.Any(k, v))
		}

		// TODO: this is a hack to get rid of the extended fields every time we log
		// log wrapper is not correctly cloned
		// here use zap.Logger
		lg := l.L.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			lg.Debug(msg)
		case logging.LevelInfo:
			lg.Info(msg)
		case logging.LevelWarn:
			lg.Warn(msg)
		case logging.LevelError:
			lg.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
