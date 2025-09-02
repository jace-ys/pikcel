package instrument

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"

	"github.com/jace-ys/pikcel/internal/ctxlog"
)

var metrics struct {
	init                   sync.Once
	goPanicsRecoveredTotal metric.Int64Counter
}

func initMetrics(ctx context.Context) {
	metrics.init.Do(func() {
		var err error
		metrics.goPanicsRecoveredTotal, err = OTel.Meter().Int64Counter("go.panics.recovered.total")
		if err != nil {
			ctxlog.Error(ctx, "error initializing metric", err)
			metrics.goPanicsRecoveredTotal, _ = noop.Meter{}.Int64Counter("") //nolint:errcheck
		}
	})
}

func EmitRecoveredPanicTelemetry(ctx context.Context, rvr any, source string) {
	initMetrics(ctx)

	err := fmt.Errorf("%v", rvr)

	ctxlog.Error(ctx, "recovered from panic", err, ctxlog.KV("panic.source", source))
	middleware.PrintPrettyStack(rvr)

	attrs := []attribute.KeyValue{
		attribute.String("panic.source", source),
	}

	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, "recovered from panic")
	span.SetAttributes(attribute.Bool("panic.recovered", true))
	span.SetAttributes(attrs...)
	span.RecordError(err, trace.WithStackTrace(true))

	metrics.goPanicsRecoveredTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
}
