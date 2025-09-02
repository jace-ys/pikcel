package telemetry

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func HTTP(attrs ...attribute.KeyValue) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			operation := r.Method + " " + r.URL.Path

			opts := []otelhttp.Option{
				otelhttp.WithSpanOptions(trace.WithAttributes(attrs...)),
			}

			handler := otelhttp.NewHandler(otelhttp.WithRouteTag(r.URL.Path, next), operation, opts...)
			handler.ServeHTTP(w, r)
		})
	}
}
