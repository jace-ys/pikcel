package recovery

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"goa.design/clue/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jace-ys/pikcel/internal/instrument"
)

func UnaryServerInterceptor(logCtx context.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (_ any, err error) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// ctx := log.WithContext(ctx, logCtx)
				instrument.EmitRecoveredPanicTelemetry(ctx, rvr, strings.TrimPrefix(info.FullMethod, "/"))
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return next(ctx, req)
	}
}

func HTTP(logCtx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if err, ok := rvr.(error); ok && errors.Is(err, http.ErrAbortHandler) {
						panic(rvr)
					}

					ctx := log.WithContext(r.Context(), logCtx)
					source := r.Method + " " + r.URL.Path
					instrument.EmitRecoveredPanicTelemetry(ctx, rvr, source)

					if r.Header.Get("Connection") != "Upgrade" {
						w.WriteHeader(http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
