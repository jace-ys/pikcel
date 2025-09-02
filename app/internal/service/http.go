package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"goa.design/clue/debug"
	"goa.design/clue/log"
	"goa.design/goa/v3/http/middleware"

	"github.com/jace-ys/pikcel/internal/ctxlog"
	"github.com/jace-ys/pikcel/internal/healthz"
	"github.com/jace-ys/pikcel/internal/transport/middleware/recovery"
	"github.com/jace-ys/pikcel/internal/transport/middleware/reqid"
	"github.com/jace-ys/pikcel/internal/transport/middleware/telemetry"
)

type HTTPServer struct {
	name string
	addr string
	srv  *http.Server
	mux  *chi.Mux
}

func NewHTTPServer(_ context.Context, name string, port int) *HTTPServer {
	addr := fmt.Sprintf(":%d", port)

	return &HTTPServer{
		name: name,
		addr: addr,
		srv: &http.Server{
			Addr:              addr,
			ReadHeaderTimeout: time.Second,
		},
		mux: chi.NewRouter(),
	}
}

func (s *HTTPServer) RegisterHandler(h http.Handler) {
	s.mux.Mount("/", h)
}

var _ Server = (*HTTPServer)(nil)

func (s *HTTPServer) Name() string {
	return s.name
}

func (s *HTTPServer) Kind() string {
	return "http"
}

func (s *HTTPServer) Addr() string {
	return s.addr
}

func (s *HTTPServer) Serve(ctx context.Context) error {
	s.srv.Handler = s.router(ctx)
	if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serving HTTP server: %w", err)
	}
	return nil
}

func (s *HTTPServer) router(ctx context.Context) http.Handler {
	s.mux.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	excludedPaths := map[string]bool{
		"/healthz": true,
	}

	logCtx := log.With(ctx, ctxlog.KV("server", s.Name()))
	return chainMiddleware(s.mux,
		withPathFilter(telemetry.HTTP(attribute.String("http.server.name", s.Name())), excludedPaths),
		recovery.HTTP(logCtx),
		withPathFilter(middleware.PopulateRequestContext(), excludedPaths),
		withPathFilter(reqid.HTTP(), excludedPaths),
		withPathFilter(ctxlog.HTTP(logCtx), excludedPaths),
		withPathFilter(debug.HTTP(), excludedPaths),
	)
}

func chainMiddleware(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func withPathFilter(m func(http.Handler) http.Handler, excluded map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exclude := excluded[r.URL.Path]; exclude {
				next.ServeHTTP(w, r)
				return
			}
			m(next).ServeHTTP(w, r)
		})
	}
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx) //nolint:wrapcheck
}

var _ healthz.Target = (*HTTPServer)(nil)

func (s *HTTPServer) HealthChecks() []health.Check {
	return []health.Check{
		healthz.HTTPCheck(s.Name(), fmt.Sprintf("http://%s/healthz", s.Addr())),
	}
}
