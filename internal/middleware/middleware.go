package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// LoggingInterceptor returns a unary server interceptor that logs the gRPC method,
// request, and duration using the provided Zap logger.
func LoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		logger.Info("gRPC request",
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		logger.Info("gRPC response",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Any("error", err),
		)
		return resp, err
	}
}

// ChainUnaryServer chains multiple unary server interceptors.
func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	n := len(interceptors)
	if n == 0 {
		return nil
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chainer := func(currentInter grpc.UnaryServerInterceptor, currentHandler grpc.UnaryHandler) grpc.UnaryHandler {
			return func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInter(currentCtx, currentReq, info, currentHandler)
			}
		}

		chainedHandler := handler
		for i := n - 1; i >= 0; i-- {
			chainedHandler = chainer(interceptors[i], chainedHandler)
		}
		return chainedHandler(ctx, req)
	}
}
