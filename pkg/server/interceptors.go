package server

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	uuid "github.com/satori/go.uuid"
)

// RequestID exracts the request id from context, if there is
// no request id a fresh ID is generated
func RequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.NewV4().String()
	}
	if v, ok := md["request_id"]; ok && len(v) > 0 {
		return v[0]
	}
	return uuid.NewV4().String()
}

// LogUnaryInterceptor logs unary method calls
func LogUnaryInterceptor(log zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var start = time.Now().UTC()
		log = log.With().Fields(map[string]interface{}{
			"request_id":  RequestID(ctx),
			"grpc_method": info.FullMethod,
		}).Logger()
		ctx = log.WithContext(ctx)
		defer log.Debug().
			TimeDiff("gprc_request_duration", time.Now().UTC(), start).
			Msg("handled gRPC unary request")
		return handler(ctx, req)
	}
}

// WrappedServerStream is a thin wrapper around grpc.ServerStream
// that allows modifying context
type WrappedServerStream struct {
	grpc.ServerStream

	WrappedContext context.Context
}

// Context returns the wrapper's WrappedContext,
// overwriting the nested grpc.ServerStream.Context()
func (w *WrappedServerStream) Context() context.Context {
	return w.WrappedContext
}

// LogStreamInyerceptor logs stream method calls
func LogStreamInyerceptor(log zerolog.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var start = time.Now().UTC()
		log = log.With().Fields(map[string]interface{}{
			"request_id":  RequestID(ss.Context()),
			"grpc_method": info.FullMethod,
		}).Logger()
		ctx := log.WithContext(ss.Context())
		ws := &WrappedServerStream{ss, ctx}
		defer log.Debug().
			TimeDiff("gprc_request_duration", time.Now().UTC(), start).
			Msg("handled gRPC stream request")
		return handler(srv, ws)
	}
}
