package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/core/log"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func codeToLevel(code codes.Code) log.Level {
	switch code {
	case codes.OK:
		return log.InfoLevel // log.DebugLevel
	case codes.NotFound, codes.Canceled, codes.AlreadyExists, codes.InvalidArgument, codes.Unauthenticated:
		return log.InfoLevel

	case codes.DeadlineExceeded, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition,
		codes.Aborted, codes.OutOfRange, codes.Unavailable:
		return log.WarnLevel

	case codes.Unknown, codes.Unimplemented, codes.Internal, codes.DataLoss:
		return log.ErrorLevel

	default:
		return log.WarnLevel
	}
}

// SimpleInterceptorLogger is a simplified gRPC server request logger. For an
// alternative, see the example from
// go-grpc-middleware/interceptors/logging/examples/zap/example_test.go
func SimpleInterceptorLogger(l *log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		tStart := time.Now()
		resp, err := handler(ctx, req)
		elapsedMs := float64(time.Since(tStart).Microseconds()) / 1e3
		code := status.Code(err)
		fields := []zap.Field{
			zap.String("method", strings.Trim(info.FullMethod, "/")),
			zap.String("elapsed", fmt.Sprintf("%.3fms", elapsedMs)),
			zap.String("code", code.String()),
		}
		var msg string
		if err != nil {
			msg = "call failure"
			fields = append(fields, zap.Error(err))
		} else {
			msg = "call success"
		}
		l.Log(codeToLevel(code), msg, fields...)
		return resp, err
	}
}
