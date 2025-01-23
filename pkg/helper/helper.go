package helper

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type Api struct {
	Logger      *slog.Logger
	Directories DirectoryMap
}

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
// This is called automatically if [Helper.Main] is called.  Otherwise, it is expected that this function is called prior to calling any [Helper] methods.
// Returns an error if invalid arguments are provided to the [Helper].
func (h *Helper) initialize() error {
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
func (h *Helper) runCallback(cb Callback) error {
	api := Api{Directories: DirectoryMap(h.Directories), Logger: h.Logger}
	return cb(h.Context, api)
}

// Runs the autopublisher.
// Returns a failure if the attempt to autopublish fails.
func (h *Helper) autopublish(ctx context.Context, api Api) error {
	return nil
}

// 'Bootstraps' the entrypoint.
// When run as root, will determine a non-root user, take ownership of necessary directories with this non-root user, and then relaunch the entrypoint as this non-root user.
// When run as non-root, will directly launch the entrypoint as the non-root user.
func (h *Helper) bootstrap(ctx context.Context, api Api) error {
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

		err = api.SetOwnerForPaths(runAsUser, api.Directories.List()...)
		if err != nil {
			return err
		}
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	_, err = api.RunCommand([]string{executable, "entrypoint"}, CmdOpts{Attach: true, Env: os.Environ(), User: runAsUser})
	return err
}

// Runs the entrypoint.
// Returns an error if the entrypoint action fails.
func (h *Helper) entrypoint(ctx context.Context, api Api) error {
	return h.Entrypoint(ctx, api)
}

// Runs the health check.
// Returns an error if the health check fails.
func (h *Helper) healthCheck(ctx context.Context, api Api) error {
	if h.HealthCheck == nil {
		return fmt.Errorf("health check not defined")
	}

	return h.runCallback(h.HealthCheck)
}

// Prints the version
func (h *Helper) printVersion(ctx context.Context, api Api) error {
	fmt.Print(h.Version)
	return nil
}

// Runs the helper with the provided arguments.
// Returns an error on failure.
func (h *Helper) main(args ...string) error {
	err := h.initialize()
	if err != nil {
		return err
	}

	cmd := "bootstrap"
	if len(args) >= 2 {
		cmd = args[1]
	}

	switch cmd {
	case "autopublish":
		err = h.runCallback(h.autopublish)
	case "bootstrap":
		err = h.runCallback(h.bootstrap)
	case "entrypoint":
		err = h.runCallback(h.entrypoint)
	case "health":
		err = h.runCallback(h.healthCheck)
	case "version":
		err = h.runCallback(h.printVersion)
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
