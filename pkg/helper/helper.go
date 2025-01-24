package helper

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/benfiola/game-server-helper/pkg/helperapi"
)

// A callback is an helper task ultimately invoked through [Helper.runCallback]
type Callback func(ctx context.Context, api Api) error

// An Helper wraps common tasks that need to be performed by many game server docker images.
type Helper struct {
	Autopublisher Autopublisher
	Context       context.Context
	Directories   map[string]string
	Entrypoint    Callback
	HealthCheck   Callback
	Logger        *slog.Logger
	Version       string
}

// Initialies the helper - setting struct member defaults and validating others.
// This is called automatically if [Helper.Run] is called.  Otherwise, it is expected that this function is called prior to calling any [Helper] methods.
// Returns an error if invalid arguments are provided to the [Helper].
func (h *Helper) Initialize() error {
	if h.Context == nil {
		h.Context = context.Background()
	}
	if h.Directories == nil {
		h.Directories = map[string]string{}
	}
	if h.Entrypoint == nil {
		return fmt.Errorf("entrypoint must be defined")
	}
	if h.Logger == nil {
		h.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	}
	if h.Version == "" {
		return fmt.Errorf("version must be defined")
	}
	return nil
}

// Creates a [context.Context] and runs the given [callback] with it.
func (h *Helper) RunCallback(cb Callback) error {
	api := Api{Api: helperapi.Api{Logger: h.Logger}, Directories: Map[string, string](h.Directories)}
	return cb(h.Context, api)
}

// Runs the helper with the provided arguments.
// Returns an error on failure.
func (h *Helper) main(args ...string) error {
	err := h.Initialize()
	if err != nil {
		return err
	}

	cmd := "bootstrap"
	if len(args) >= 2 {
		cmd = args[1]
	}

	var callback Callback
	switch cmd {
	case "autopublish":
		callback = h.Autopublish
	case "bootstrap":
		callback = h.Bootstrap
	case "entrypoint":
		callback = h.EntrypointCallback
	case "health":
		callback = h.HealthCallback
	case "version":
		callback = h.PrintVersion
	default:
		return fmt.Errorf("unknown command %s", cmd)
	}

	return h.RunCallback(callback)
}

// Runs the helper with the process arguments, and exits on completion.
// Exits with status code 0 on success.
// Exits with status code 1 on failurh.
func (h *Helper) Run() {
	err := h.main(os.Args...)

	code := 0
	if err != nil {
		code = 1
		h.Logger.Error("helper failed", "error", err.Error())
	}

	os.Exit(code)
}
