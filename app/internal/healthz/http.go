package healthz

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexliesenfeld/health"
)

func HTTPCheck(name, url string) health.Check {
	return health.Check{
		Name: "http:" + name,
		Check: func(ctx context.Context) error {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("create HTTP request: %w", err)
			}
			req.Header.Set("Connection", "close")

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("send HTTP request: %w", err)
			}
			defer res.Body.Close()

			if res.StatusCode >= http.StatusInternalServerError {
				return fmt.Errorf("HTTP service reported as non-healthy: %d", res.StatusCode)
			}

			return nil
		},
	}
}
