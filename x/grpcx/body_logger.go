package grpcx

import (
	"context"
	"kwil/x/logx"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// BodyLoggerInterceptor returns a new unary server interceptor that logs the
// request body.
func BodyLoggerInterceptor(logger logx.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Debug("request body", zap.Any("body", req))
		return handler(ctx, req)
	}
}
