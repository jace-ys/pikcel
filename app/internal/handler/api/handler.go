package api

import (
	"context"

	"github.com/alexliesenfeld/health"

	"github.com/jace-ys/pikcel/api/v1/gen/api"
	"github.com/jace-ys/pikcel/internal/healthz"
	"github.com/jace-ys/pikcel/internal/idgen"
)

var _ api.Service = (*Handler)(nil)

type Handler struct{}

func NewHandler() (*Handler, error) {
	return &Handler{}, nil
}

func (h *Handler) CanvasGet(_ context.Context) (*api.Canvas, error) {
	return &api.Canvas{
		ID:     idgen.New[idgen.Canvas]().String(),
		Width:  100,
		Height: 100,
	}, nil
}

var _ healthz.Target = (*Handler)(nil)

func (h *Handler) HealthChecks() []health.Check {
	return []health.Check{
		{
			Name: "handler:api",
			Check: func(_ context.Context) error {
				return nil
			},
		},
	}
}
