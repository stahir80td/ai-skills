package reliability

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCClientConfig configures the gRPC client with reliability patterns
type GRPCClientConfig struct {
	Name               string
	CircuitMaxFailures uint32
	CircuitTimeout     time.Duration
	RateLimit          float64
	RateBurst          int
	MaxConcurrency     int
	BulkheadTimeout    time.Duration
	EnableRetry        bool
	RetryConfig        RetryConfig
}

// DefaultGRPCClientConfig returns sensible defaults for gRPC client
func DefaultGRPCClientConfig(name string) GRPCClientConfig {
	return GRPCClientConfig{
		Name:               name,
		CircuitMaxFailures: 5,
		CircuitTimeout:     60 * time.Second,
		RateLimit:          100,
		RateBurst:          200, // Increased for burst traffic
		MaxConcurrency:     50,
		BulkheadTimeout:    5 * time.Second,
		EnableRetry:        true,
		RetryConfig:        DefaultRetryConfig(),
	}
}

// UnaryClientInterceptor wraps gRPC unary calls with circuit breaker + retry + rate limiting
func UnaryClientInterceptor(config GRPCClientConfig) grpc.UnaryClientInterceptor {
	cb := NewCircuitBreaker(config.Name+"-circuit", config.CircuitMaxFailures, config.CircuitTimeout)
	rateLimiter := NewRateLimiter(config.Name+"-rate", config.RateLimit, config.RateBurst)
	bulkhead := NewBulkhead(config.Name+"-bulkhead", config.MaxConcurrency, config.BulkheadTimeout)

	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		// Rate limiting
		if !rateLimiter.Allow() {
			return status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		// Bulkhead + Circuit Breaker + Retry
		return bulkhead.ExecuteWithContext(ctx, func(ctx context.Context) error {
			return cb.ExecuteWithContext(ctx, func(ctx context.Context) error {
				if config.EnableRetry {
					return Retry(ctx, config.Name+"-"+method, config.RetryConfig, func() error {
						return invoker(ctx, method, req, reply, cc, opts...)
					})
				}
				return invoker(ctx, method, req, reply, cc, opts...)
			})
		})
	}
}

// StreamClientInterceptor wraps gRPC streams with bulkhead
func StreamClientInterceptor(config GRPCClientConfig) grpc.StreamClientInterceptor {
	bulkhead := NewBulkhead(config.Name+"-stream-bulkhead", config.MaxConcurrency, config.BulkheadTimeout)

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
		method string, streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {

		var stream grpc.ClientStream
		err := bulkhead.ExecuteWithContext(ctx, func(ctx context.Context) error {
			var err error
			stream, err = streamer(ctx, desc, cc, method, opts...)
			return err
		})
		return stream, err
	}
}

// UnaryServerInterceptor wraps gRPC server unary calls with rate limiting + bulkhead
func UnaryServerInterceptor(config GRPCClientConfig) grpc.UnaryServerInterceptor {
	rateLimiter := NewRateLimiter(config.Name+"-server-rate", config.RateLimit, config.RateBurst)
	bulkhead := NewBulkhead(config.Name+"-server-bulkhead", config.MaxConcurrency, config.BulkheadTimeout)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		// Rate limiting
		if !rateLimiter.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "server rate limit exceeded")
		}

		// Bulkhead
		var resp interface{}
		err := bulkhead.ExecuteWithContext(ctx, func(ctx context.Context) error {
			var err error
			resp, err = handler(ctx, req)
			return err
		})
		return resp, err
	}
}

// StreamServerInterceptor wraps gRPC server streams with bulkhead
func StreamServerInterceptor(config GRPCClientConfig) grpc.StreamServerInterceptor {
	bulkhead := NewBulkhead(config.Name+"-server-stream-bulkhead", config.MaxConcurrency, config.BulkheadTimeout)

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) error {

		return bulkhead.Execute(func() error {
			return handler(srv, ss)
		})
	}
}
