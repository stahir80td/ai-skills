package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/your-github-org/ai-scaffolder/core/go/sli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewSLIUnaryInterceptor creates a gRPC unary interceptor for SLI metrics tracking
func NewSLIUnaryInterceptor(tracker sli.Tracker) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Record metrics
		duration := time.Since(start)

		// Extract operation name from full method (e.g., "/proto.Service/Method")
		operationName := extractOperationName(info.FullMethod)

		// Determine success/failure
		success := err == nil
		errorCode := "OK"
		errorSeverity := "low"

		if err != nil {
			st, _ := status.FromError(err)
			success = false
			errorCode = st.Code().String()

			// Map gRPC status codes to severity levels
			switch st.Code() {
			case codes.Internal, codes.Unknown:
				errorSeverity = "critical"
			case codes.Unavailable, codes.DeadlineExceeded:
				errorSeverity = "high"
			case codes.PermissionDenied, codes.Unauthenticated:
				errorSeverity = "medium"
			case codes.InvalidArgument, codes.NotFound:
				errorSeverity = "low"
			}
		}

		// Record SLI metrics
		outcome := sli.RequestOutcome{
			Operation:     operationName,
			Success:       success,
			Latency:       duration,
			ErrorCode:     errorCode,
			ErrorSeverity: errorSeverity,
		}

		tracker.RecordRequest(ctx, outcome)

		return resp, err
	}
}

// NewSLIStreamInterceptor creates a gRPC stream interceptor for SLI metrics tracking
func NewSLIStreamInterceptor(tracker sli.Tracker) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Call the handler
		err := handler(srv, ss)

		// Record metrics
		duration := time.Since(start)
		operationName := extractOperationName(info.FullMethod)

		success := err == nil
		errorCode := "OK"
		errorSeverity := "low"

		if err != nil {
			st, _ := status.FromError(err)
			success = false
			errorCode = st.Code().String()

			switch st.Code() {
			case codes.Internal, codes.Unknown:
				errorSeverity = "critical"
			case codes.Unavailable, codes.DeadlineExceeded:
				errorSeverity = "high"
			case codes.PermissionDenied, codes.Unauthenticated:
				errorSeverity = "medium"
			case codes.InvalidArgument, codes.NotFound:
				errorSeverity = "low"
			}
		}

		outcome := sli.RequestOutcome{
			Operation:     operationName,
			Success:       success,
			Latency:       duration,
			ErrorCode:     errorCode,
			ErrorSeverity: errorSeverity,
		}

		tracker.RecordRequest(context.Background(), outcome)

		return err
	}
}

// ChainUnaryInterceptors chains multiple unary interceptors in order
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Build chain from right to left (last interceptor wraps first)
		var chain grpc.UnaryHandler = handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			chain = func(currentInterceptor grpc.UnaryServerInterceptor, nextHandler grpc.UnaryHandler) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return currentInterceptor(ctx, req, info, nextHandler)
				}
			}(interceptor, chain)
		}
		return chain(ctx, req)
	}
}

// ChainStreamInterceptors chains multiple stream interceptors in order
func ChainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Build chain from right to left (last interceptor wraps first)
		var chain grpc.StreamHandler = handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			chain = func(currentInterceptor grpc.StreamServerInterceptor, nextHandler grpc.StreamHandler) grpc.StreamHandler {
				return func(srv interface{}, ss grpc.ServerStream) error {
					return currentInterceptor(srv, ss, info, nextHandler)
				}
			}(interceptor, chain)
		}
		return chain(srv, ss)
	}
}

// extractOperationName extracts a clean operation name from gRPC full method
// e.g., "/proto.Service/Method" -> "Method"
func extractOperationName(fullMethod string) string {
	// Full method format: /package.Service/Method
	parts := strings.Split(fullMethod, "/")
	if len(parts) >= 3 {
		return parts[len(parts)-1] // Return the method name
	}
	return fullMethod
}
