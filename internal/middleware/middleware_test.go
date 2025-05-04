package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestLoggingInterceptor(t *testing.T) {
	logger, _ := zap.NewProduction()
	interceptor := LoggingInterceptor(logger)

	ctx := context.Background()
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Method",
	}

	_, err := interceptor(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return "test response", nil
	})

	assert.NoError(t, err, "Interceptor should not return an error")
	assert.NotNil(t, logger, "Logger should not be nil")
}