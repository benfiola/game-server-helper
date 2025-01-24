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
	Context     context.Context
	Directories map[string]string
	Entrypoint  Callback
	HealthCheck Callback
	Logger      *slog.Logger
	Version     string
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

// 'Bootstraps' the entrypoint.
// When run as root, will determine a non-root user, take ownership of necessary directories with this non-root user, and then relaunch the entrypoint as this non-root user.
// When run as non-root, will directly launch the entrypoint as the non-root user.
func (h *Helper) Bootstrap(ctx context.Context, api Api) error {
	api.Logger.Info("bootstrap")

	runAsUser, err := api.GetCurrentUser()
	if err != nil {
		return err
	}
	currentUser := runAsUser

	if currentUser.Uid == 0 {
		runAsUser, err = api.GetEnvUser()
		if err != nil {
			return err
		}

		err = api.UpdateUser("server", runAsUser)
		if err != nil {
			return err
		}

		err = api.SetOwnerForPaths(runAsUser, api.Directories.Values()...)
		if err != nil {
			return err
		}
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	_, err = api.RunCommand([]string{executable, "entrypoint"}, helperapi.CmdOpts{Attach: true, Env: os.Environ(), User: runAsUser})
	return err
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

	switch cmd {
	case "bootstrap":
		err = h.RunCallback(h.Bootstrap)
	case "entrypoint":
		err = h.RunCallback(h.Entrypoint)
	case "health":
		err = fmt.Errorf("health check not defined")
		if h.HealthCheck != nil {
			err = h.RunCallback(h.HealthCheck)
		}
	case "version":
		fmt.Print(h.Version)
	default:
		err = fmt.Errorf("unknown command %s", cmd)
	}

	return err
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
