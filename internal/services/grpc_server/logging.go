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
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func codeToLevel(code codes.Code) log.Level {
	switch code {
	case codes.OK:
		return log.InfoLevel // log.DebugLevel
	case codes.NotFound, codes.Canceled, codes.AlreadyExists, codes.InvalidArgument, codes.Unauthenticated:
		return log.InfoLevel

	case codes.DeadlineExceeded, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition,
		codes.Aborted, codes.OutOfRange, codes.Unavailable, codes.Unknown, codes.Unimplemented:
		return log.WarnLevel

	case codes.Internal, codes.DataLoss:
		// WARNING: This error level will result in a call stack dump, so try
		// not to use these codes unless we know that there is a server error,
		// as opposed to a user-generated error such as bad inputs. The docs for
		// Internal state that it should indicate "something is very broken".
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
		fields := []zap.Field{
			zap.String("method", strings.Trim(info.FullMethod, "/")),
			zap.String("elapsed", fmt.Sprintf("%.3fms", elapsedMs)),
		}
		if peer, ok := peer.FromContext(ctx); ok {
			fields = append(fields, zap.String("addr", peer.Addr.String()))
		}
		stat := status.Convert(err)
		code := stat.Code()
		fields = append(fields, zap.String("code", code.String()))
		var msg string
		if err != nil {
			msg = "call failure"
		} else {
			msg = "call success"
		}
		if errDesc := stat.Message(); errDesc != "" {
			fields = append(fields, zap.String("err", errDesc))
		}
		l.Log(codeToLevel(code), msg, fields...)
		return resp, err
	}
}
