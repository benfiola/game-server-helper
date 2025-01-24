package helper

import (
	"context"
)

// Runs the entrypoint
func (h *Helper) EntrypointCallback(ctx context.Context, api Api) error {
	return h.Entrypoint(ctx, api)
}
