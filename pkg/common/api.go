package common

import "log/slog"

// Api exposes a common, composable api with utility functions
type Api struct {
	Logger *slog.Logger
}
