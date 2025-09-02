package goa

import (
	goagrpc "goa.design/goa/v3/grpc"

	"github.com/jace-ys/pikcel/internal/endpoint"
)

type GRPCAdapter[E endpoint.GoaEndpoints, S any] struct {
	newFn GRPCNewFunc[E, S]
}

type GRPCNewFunc[E endpoint.GoaEndpoints, S any] func(e E, uh goagrpc.UnaryHandler) *S

func GRPC[E endpoint.GoaEndpoints, S any](newFn GRPCNewFunc[E, S]) *GRPCAdapter[E, S] {
	return &GRPCAdapter[E, S]{
		newFn: newFn,
	}
}

func (a *GRPCAdapter[E, S]) Adapt(ep E) *S {
	return a.newFn(ep, nil)
}
