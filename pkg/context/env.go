package context

import (
	"github.com/caarlos0/env/v11"
)

// Parses environment variables into the provided struct.
// Returns an error if parsing the environment variables fail.
// See: [env.Parse]
func (ctx *Context) ParseEnv(cfg interface{}) error {
	return env.Parse(&cfg)
}
