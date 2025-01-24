package helper

import (
	"context"
	"fmt"
)

// Prints the version
func (h *Helper) HealthCallback(ctx context.Context, api Api) error {
	api.Logger.Info("health")
	if h.HealthCheck == nil {
		return fmt.Errorf("health check not configured")
	}
	return h.HealthCheck(ctx, api)
}
