package service

import (
	"context"
	"fmt"
	"net"

	"github.com/alexliesenfeld/health"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"goa.design/clue/debug"
	"goa.design/clue/log"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/stats"

	"github.com/jace-ys/pikcel/internal/ctxlog"
	"github.com/jace-ys/pikcel/internal/healthz"
	"github.com/jace-ys/pikcel/internal/transport/middleware/recovery"
	"github.com/jace-ys/pikcel/internal/transport/middleware/reqid"
)

type GRPCServer struct {
	name string
	addr string
	srv  *grpc.Server
}

func NewGRPCServer[SS any](ctx context.Context, name string, port int) *GRPCServer {
	addr := fmt.Sprintf(":%d", port)

	excludedMethods := map[string]bool{
		grpc_reflection_v1.ServerReflection_ServerReflectionInfo_FullMethodName:      true,
		grpc_reflection_v1alpha.ServerReflection_ServerReflectionInfo_FullMethodName: true,
		healthpb.Health_Check_FullMethodName:                                         true,
		healthpb.Health_Watch_FullMethodName:                                         true,
	}

	logCtx := log.With(ctx, ctxlog.KV("server", name))
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(logCtx),
			withMethodFilter(reqid.UnaryServerInterceptor(), excludedMethods),
			withMethodFilter(ctxlog.UnaryServerInterceptor(logCtx), excludedMethods),
			withMethodFilter(debug.UnaryServerInterceptor(), excludedMethods),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithSpanAttributes(attribute.String("rpc.server.name", name)),
			otelgrpc.WithFilter(func(info *stats.RPCTagInfo) bool {
				return !excludedMethods[info.FullMethodName]
			}),
		)),
	)

	reflection.Register(srv)
	healthpb.RegisterHealthServer(srv, grpchealth.NewServer())

	return &GRPCServer{
		name: name,
		addr: addr,
		srv:  srv,
	}
}

func withMethodFilter(interceptor grpc.UnaryServerInterceptor, excluded map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if exclude := excluded[info.FullMethod]; exclude {
			return handler(ctx, req)
		}
		return interceptor(ctx, req, info, handler)
	}
}

func (s *GRPCServer) RegisterHandler(sd *grpc.ServiceDesc, ss any) {
	s.srv.RegisterService(sd, ss)
}

var _ Server = (*GRPCServer)(nil)

func (s *GRPCServer) Name() string {
	return s.name
}

func (s *GRPCServer) Kind() string {
	return "grpc"
}

func (s *GRPCServer) Addr() string {
	return s.addr
}

func (s *GRPCServer) Serve(ctx context.Context) error {
	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "tcp", s.addr)
	if err != nil {
		return fmt.Errorf("tcp listener: %w", err)
	}

	if err := s.srv.Serve(lis); err != nil {
		return fmt.Errorf("serving gRPC server: %w", err)
	}

	return nil
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	ok := make(chan struct{})

	go func() {
		s.srv.GracefulStop()
		close(ok)
	}()

	select {
	case <-ok:
		return nil
	case <-ctx.Done():
		s.srv.Stop()
		return ctx.Err()
	}
}

var _ healthz.Target = (*GRPCServer)(nil)

func (s *GRPCServer) HealthChecks() []health.Check {
	return []health.Check{
		healthz.GRPCCheck(s.Name(), s.Addr()),
	}
}
