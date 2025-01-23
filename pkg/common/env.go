package common

import (
	"github.com/caarlos0/env/v11"
)

// Parses environment variables into the provided struct pointer.
// Returns an error if parsing the environment variables fail.
// See: [env.Parse]
func (api *Api) ParseEnv(cfg any) error {
	return env.Parse(cfg)
}
