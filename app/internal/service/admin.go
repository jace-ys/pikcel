package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/alexliesenfeld/health"
	"github.com/go-chi/chi/v5"
	"goa.design/clue/debug"
	"goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/http/middleware"

	"github.com/jace-ys/pikcel/internal/ctxlog"
	"github.com/jace-ys/pikcel/internal/healthz"
	"github.com/jace-ys/pikcel/internal/transport/middleware/recovery"
)

type AdminServer struct {
	debug  bool
	srv    *HTTPServer
	mux    *chi.Mux
	checks []health.Check
}

func NewAdminServer(ctx context.Context, port int, debug bool) *AdminServer {
	return &AdminServer{
		debug: debug,
		srv:   NewHTTPServer(ctx, "admin", port),
		mux:   chi.NewRouter(),
	}
}

var _ Server = (*AdminServer)(nil)

func (s *AdminServer) Name() string {
	return s.srv.Name()
}

func (s *AdminServer) Kind() string {
	return s.srv.Kind()
}

func (s *AdminServer) Addr() string {
	return s.srv.Addr()
}

func (s *AdminServer) Serve(ctx context.Context) error {
	s.srv.srv.Handler = s.router(ctx)
	if err := s.srv.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serving admin server: %w", err)
	}
	return nil
}

func (s *AdminServer) router(ctx context.Context) http.Handler {
	s.mux.Get("/healthz", health.NewHandler(healthz.NewChecker(s.checks...)))

	debugMux := goahttp.NewMuxer()
	debug.MountPprofHandlers(debug.Adapt(debugMux), debug.WithPrefix("/pprof"))
	if s.debug {
		debug.MountDebugLogEnabler(debug.Adapt(debugMux), debug.WithPath("/settings"))
	}
	s.mux.Mount("/debug", debugMux)

	logCtx := log.With(ctx, ctxlog.KV("server", s.Name()))
	return chainMiddleware(s.mux,
		recovery.HTTP(logCtx),
		middleware.PopulateRequestContext(),
	)
}

func (s *AdminServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx) //nolint:wrapcheck
}

func (s *AdminServer) Administer(targets ...healthz.Target) {
	for _, target := range targets {
		s.checks = append(s.checks, target.HealthChecks()...)
	}
}
