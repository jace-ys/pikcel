package ctxlog

import (
	"context"
	"net/http"

	"goa.design/clue/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/jace-ys/pikcel/internal/transport/middleware/reqid"
)

func UnaryServerInterceptor(logCtx context.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		requestID := reqid.RequestIDFromContext(ctx)

		ctx = log.WithContext(ctx, logCtx)
		ctx = log.With(ctx, KV(log.RequestIDKey, requestID))

		opts := []log.GRPCLogOption{
			log.WithErrorFunc(func(_ codes.Code) bool { return false }),
			log.WithDisableCallID(),
		}

		return log.UnaryServerInterceptor(ctx, opts...)(ctx, req, info, next)
	}
}

func HTTP(logCtx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			requestID := reqid.RequestIDFromContext(ctx)

			ctx = log.WithContext(ctx, logCtx)
			ctx = log.With(ctx, KV(log.RequestIDKey, requestID))

			opts := []log.HTTPLogOption{
				log.WithDisableRequestID(),
			}

			handler := log.HTTP(ctx, opts...)(next)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
