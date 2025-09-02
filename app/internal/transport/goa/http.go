package goa

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/http/middleware"
	goa "goa.design/goa/v3/pkg"

	"github.com/jace-ys/pikcel/internal/ctxlog"
	"github.com/jace-ys/pikcel/internal/endpoint"
)

type HTTPServer interface {
	MethodNames() []string
	Mount(mux goahttp.Muxer)
	Service() string
	Use(m func(http.Handler) http.Handler)
}

type HTTPAdapter[E endpoint.GoaEndpoints, S HTTPServer] struct {
	newFn   HTTPNewFunc[E, S]
	mountFn HTTPMountFunc[S]
}

type HTTPNewFunc[E endpoint.GoaEndpoints, S HTTPServer] func(
	e E,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	httpFS http.FileSystem,
) S

type HTTPMountFunc[S HTTPServer] func(mux goahttp.Muxer, srv S)

func HTTP[E endpoint.GoaEndpoints, S HTTPServer](newFn HTTPNewFunc[E, S], mountFn HTTPMountFunc[S]) *HTTPAdapter[E, S] {
	return &HTTPAdapter[E, S]{
		newFn:   newFn,
		mountFn: mountFn,
	}
}

func (a *HTTPAdapter[E, S]) Adapt(ep E, fsys fs.FS) goahttp.ResolverMuxer {
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	formatter := goahttp.NewErrorResponse

	eh := func(ctx context.Context, w http.ResponseWriter, err error) {
		ctxlog.Error(ctx, "failed to encode response", err,
			ctxlog.KV("http.method", ctx.Value(middleware.RequestMethodKey)),
			ctxlog.KV("http.path", ctx.Value(middleware.RequestPathKey)),
		)

		gerr := goa.Fault("failed to encode response")

		span := trace.SpanFromContext(ctx)
		span.SetStatus(codes.Error, gerr.GoaErrorName())
		span.SetAttributes(attribute.String("error", fmt.Sprintf("failed to encode response: %v", err)))

		if err := goahttp.ErrorEncoder(enc, formatter)(ctx, w, gerr); err != nil {
			panic(err)
		}
	}

	mux := goahttp.NewMuxer()
	srv := a.newFn(ep, mux, dec, enc, eh, formatter, http.FS(fsys))
	a.mountFn(mux, srv)

	return mux
}
