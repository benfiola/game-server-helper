package helper

import (
	"context"
	"fmt"
)

// Prints the version
func (h *Helper) PrintVersion(ctx context.Context, api Api) error {
	fmt.Print(h.Version)
	return nil
}
