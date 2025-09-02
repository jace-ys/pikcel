package main

import (
	"context"
	"fmt"

	apiv1 "github.com/jace-ys/pikcel/api/v1"
	genapi "github.com/jace-ys/pikcel/api/v1/gen/api"
	apipb "github.com/jace-ys/pikcel/api/v1/gen/grpc/api/pb"
	grpcapi "github.com/jace-ys/pikcel/api/v1/gen/grpc/api/server"
	httpapi "github.com/jace-ys/pikcel/api/v1/gen/http/api/server"
	"github.com/jace-ys/pikcel/internal/ctxlog"
	"github.com/jace-ys/pikcel/internal/endpoint"
	"github.com/jace-ys/pikcel/internal/handler/api"
	"github.com/jace-ys/pikcel/internal/instrument"
	"github.com/jace-ys/pikcel/internal/service"
	goatransport "github.com/jace-ys/pikcel/internal/transport/goa"
)

type ServerCmd struct {
	Port      int `default:"8080" env:"PORT" help:"Port to listen on for the HTTP server."`
	AdminPort int `default:"9090" env:"ADMIN_PORT" help:"Port to listen on for the admin server."`
}

func (c *ServerCmd) Run(ctx context.Context, g *Globals) error {
	if err := instrument.InitOTel(ctx, genapi.APIName, genapi.APIVersion); err != nil {
		return fmt.Errorf("init otel instrumentation: %w", err)
	}
	defer func() {
		if err := instrument.OTel.Shutdown(ctx); err != nil {
			ctxlog.Error(ctx, "error shutting down otel provider", err)
		}
	}()

	httpSrv := service.NewHTTPServer(ctx, "pikcel", c.Port)
	grpcSrv := service.NewGRPCServer[apipb.APIServer](ctx, "pikcel", c.Port+1)

	adminSrv := service.NewAdminServer(ctx, c.AdminPort, g.Debug)
	adminSrv.Administer(httpSrv, grpcSrv)

	handler, err := api.NewHandler()
	if err != nil {
		return fmt.Errorf("init api handler: %w", err)
	}
	adminSrv.Administer(handler)

	ep := endpoint.Goa(genapi.NewEndpoints).Adapt(handler)

	{
		transport := goatransport.HTTP(httpapi.New, httpapi.Mount)
		httpSrv.RegisterHandler(transport.Adapt(ep, apiv1.OpenAPIFS))
	}
	{
		transport := goatransport.GRPC(grpcapi.New)
		grpcSrv.RegisterHandler(&apipb.API_ServiceDesc, transport.Adapt(ep))
	}

	if err := service.New(httpSrv, grpcSrv, adminSrv).Run(ctx); err != nil {
		ctxlog.Error(ctx, "encountered error while running service", err)
		return fmt.Errorf("service run: %w", err)
	}

	return nil
}
