package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	goa "goa.design/goa/v3/pkg"

	"github.com/jace-ys/pikcel/internal/instrument"
)

func Endpoint(e goa.Endpoint) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		service := "unknown"
		if s, ok := ctx.Value(goa.ServiceKey).(string); ok {
			service = s
		}

		method := "Unknown"
		if m, ok := ctx.Value(goa.MethodKey).(string); ok {
			method = m
		}

		source := fmt.Sprintf("goa.endpoint/%s.%s", service, method)
		ctx, span := instrument.OTel.Tracer().Start(ctx, source)
		span.SetAttributes(attribute.String("endpoint.service", service), attribute.String("endpoint.method", method))
		defer span.End()

		return e(ctx, req)
	}
}
