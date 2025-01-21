package utils

import (
	"context"
	"log/slog"
)

type ctxKeyLogger struct{}

type Context struct {
	context.Context
}

func (c Context) Logger() *slog.Logger {
	return c.Value(ctxKeyLogger{}).(*slog.Logger)
}
