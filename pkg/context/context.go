package context

import (
	"context"
)

// Context wraps an underlying [context.Context] and exposes convenience methods.
type Context struct {
	context.Context
}
