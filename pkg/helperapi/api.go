package helperapi

import "log/slog"

// Api is exposed to callbacks, providing helper functions to perform common operations
type Api struct {
	Logger *slog.Logger
}
