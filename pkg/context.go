package helper

import (
	"context"
	"log/slog"
)

// ctxKeyDirs is a context key pointing a mapping of [name] -> path
type ctxKeyDirs struct{}

// Retrieves a mapping of [name] -> path from the given context.
func Dirs(ctx context.Context) Map[string, string] {
	return ctx.Value(ctxKeyDirs{}).(Map[string, string])
}

// ctxKeyFileCacheEnabled is a context key pointing boolean determining whether file caching is enabled
type ctxKeyFileCacheEnabled struct{}

// Retrieves a boolean indicating whether file caching is enabled
func FileCacheEnabled(ctx context.Context) bool {
	return ctx.Value(ctxKeyFileCacheEnabled{}).(bool)
}

// ctxKeyFileCacheSizeLimit is a context key pointing to a configured file cache size limit in megabytes
type ctxKeyFileCacheSizeLimit struct{}

// Retrieves a file cache size limit (in megabytes) from the given context
func FileCacheSizeLimit(ctx context.Context) int {
	return ctx.Value(ctxKeyFileCacheSizeLimit{}).(int)
}

// ctxKeyLogger is a context key pointing to a logger
type ctxKeyLogger struct{}

// Retrieves a logger from the given context
func Logger(ctx context.Context) *slog.Logger {
	return ctx.Value(ctxKeyLogger{}).(*slog.Logger)
}

// ctxKeyUuid is a context key pointing to a session uuid
type ctxKeyUuid struct{}

// Retrieves the session uuid from the given context
func Uuid(ctx context.Context) string {
	return ctx.Value(ctxKeyUuid{}).(string)
}

// ctxKeyVersion is a context key pointing to the entrypoint version
type ctxKeyVersion struct{}

// Retrieves the entrypoint version from the given context
func Version(ctx context.Context) string {
	return ctx.Value(ctxKeyVersion{}).(string)
}
