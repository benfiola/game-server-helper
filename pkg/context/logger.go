package context

import "log/slog"

// ctxLoggerKey is used to retrieve a logger stored as a value within a [Context]
type ctxLoggerKey struct{}

// Retrieve a logger stored as a value within a [Context]
func (ctx *Context) Logger() *slog.Logger {
	return ctx.Value(ctxLoggerKey{}).(*slog.Logger)
}
