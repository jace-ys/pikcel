package reqid

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"goa.design/clue/log"
	"google.golang.org/grpc"

	"github.com/jace-ys/pikcel/internal/idgen"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		ctx = newRequestID(ctx)
		return next(ctx, req)
	}
}

func HTTP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := newRequestID(r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type ctxKeyRequestID struct{}

func newRequestID(ctx context.Context) context.Context {
	requestID := idgen.New[idgen.Request]()
	ctx = context.WithValue(ctx, ctxKeyRequestID{}, requestID)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String(log.RequestIDKey, requestID.String()))

	return ctx
}

func RequestIDFromContext(ctx context.Context) idgen.ID[idgen.Request] {
	requestID, ok := ctx.Value(ctxKeyRequestID{}).(idgen.ID[idgen.Request])
	if !ok {
		return idgen.ID[idgen.Request]{}
	}
	return requestID
}
