package service

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/jace-ys/pikcel/internal/ctxlog"
)

type Service struct {
	servers []Server
}

func New(servers ...Server) *Service {
	return &Service{
		servers: servers,
	}
}

type Server interface {
	Name() string
	Kind() string
	Addr() string
	Serve(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func (s *Service) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		stop()
	}()

	g, ctx := errgroup.WithContext(ctx)

	for _, srv := range s.servers {
		g.Go(func() error {
			ctxlog.Print(ctx, "server listening",
				ctxlog.KV("server", srv.Name()),
				ctxlog.KV("kind", srv.Kind()),
				ctxlog.KV("addr", srv.Addr()),
			)
			return srv.Serve(ctx)
		})
	}

	ctxlog.Print(ctx, "service started")
	<-ctx.Done()
	ctxlog.Print(ctx, "service shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, srv := range s.servers {
		g.Go(func() error {
			if err := srv.Shutdown(shutdownCtx); err != nil {
				ctxlog.Error(ctx, "server shutdown error", err,
					ctxlog.KV("server", srv.Name()),
					ctxlog.KV("kind", srv.Kind()),
				)
			} else {
				ctxlog.Print(ctx, "server shutdown complete",
					ctxlog.KV("server", srv.Name()),
					ctxlog.KV("kind", srv.Kind()),
				)
			}
			return nil
		})
	}

	defer ctxlog.Print(ctx, "service stopped")
	return g.Wait() //nolint:wrapcheck
}
