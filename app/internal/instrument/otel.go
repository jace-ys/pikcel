package instrument

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"goa.design/clue/clue"

	"github.com/jace-ys/pikcel/internal/versioninfo"
)

var OTel *OTelProvider

type OTelProvider struct {
	cfg           *clue.Config
	shutdownFuncs []func(context.Context) error
}

func InitOTel(ctx context.Context, name string, version string) error {
	var shutdownFuncs []func(context.Context) error

	metrics, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("metrics exporter: %w", err)
	}
	shutdownFuncs = append(shutdownFuncs, metrics.Shutdown)

	traces, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("trace exporter: %w", err)
	}
	shutdownFuncs = append(shutdownFuncs, traces.Shutdown)

	opts := []clue.Option{
		clue.WithResource(resource.Environment()),
		clue.WithReaderInterval(30 * time.Second),
	}

	cfg, err := clue.NewConfig(ctx, name, version, metrics, traces, opts...)
	if err != nil {
		return fmt.Errorf("new config: %w", err)
	}
	clue.ConfigureOpenTelemetry(ctx, cfg)

	OTel = &OTelProvider{
		cfg:           cfg,
		shutdownFuncs: shutdownFuncs,
	}

	if err := OTel.initDefaultMetrics(ctx); err != nil {
		return fmt.Errorf("default metrics: %w", err)
	}

	return nil
}

func (i *OTelProvider) initDefaultMetrics(ctx context.Context) error {
	if err := runtime.Start(); err != nil {
		return fmt.Errorf("runtime: %w", err)
	}

	up, err := OTel.Meter().Int64Gauge("up")
	if err != nil {
		return fmt.Errorf("up: %w", err)
	}
	up.Record(ctx, 1)

	return nil
}

func (i *OTelProvider) Shutdown(ctx context.Context) error {
	select {
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var errs error
	for _, shutdown := range i.shutdownFuncs {
		errs = errors.Join(errs, shutdown(shutdownCtx))
	}

	i.shutdownFuncs = nil
	return errs
}

const scope = "github.com/jace-ys/pikcel/internal/instrument"

func (i *OTelProvider) Meter() metric.Meter {
	return i.cfg.MeterProvider.Meter(scope, metric.WithInstrumentationVersion(versioninfo.Version))
}

func (i *OTelProvider) Tracer() trace.Tracer {
	return i.cfg.TracerProvider.Tracer(scope, trace.WithInstrumentationVersion(versioninfo.Version))
}

func (i *OTelProvider) Propagators() propagation.TextMapPropagator {
	return i.cfg.Propagators
}
