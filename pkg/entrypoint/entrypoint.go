package entrypoint

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/benfiola/game-server-helper/pkg/common"
)

// DirectoryMap stores a label to path mapping
type DirectoryMap map[string]string

// Returns a list of paths stored in the directory map
func (dm *DirectoryMap) List() []string {
	list := []string{}
	for _, directory := range *dm {
		list = append(list, directory)
	}
	return list
}

// Composes a [common.Api] and adds entrypoint specific information
type Api struct {
	common.Api
	Directories DirectoryMap
}

// A callback is an entrypoint task ultimately invoked through [Entrypoint.RunCallback]
type callback func(ctx context.Context, api Api) error

// An Entrypoint wraps common tasks that need to be performed by many game server docker images.
type Entrypoint struct {
	Action      callback
	Context     context.Context
	Directories map[string]string
	HealthCheck callback
	Logger      *slog.Logger
	Version     string
}

// Initialies the entrypoint - setting struct member defaults and validating others.
// This is called automatically if [Entrypoint.Main] is called.  Otherwise, it is expected that this function is called prior to calling any [Entrypoint] methods.
// Returns an error if invalid arguments are provided to the [Entrypoint].
func (e *Entrypoint) initialize() error {
	if e.Action == nil {
		return fmt.Errorf("entrypoint action must be defined")
	}
	if e.Context == nil {
		e.Context = context.Background()
	}
	if e.Directories == nil {
		e.Directories = map[string]string{}
	}
	if e.Logger == nil {
		e.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	}
	if e.Version == "" {
		return fmt.Errorf("entrypoint version must be defined")
	}
	return nil
}

// Creates a [context.Context] and runs the given [callback] with it.
func (e *Entrypoint) runCallback(cb callback) error {
	commonApi := common.Api{Logger: e.Logger}
	api := Api{Api: commonApi, Directories: DirectoryMap(e.Directories)}
	return cb(e.Context, api)
}

// Runs the entrypoint action.
// Returns an error if the entrypoint action fails.
func (e *Entrypoint) action(ctx context.Context, api Api) error {
	return e.Action(ctx, api)
}

// 'Bootstraps' the entrypoint.
// When run as root, the entrypoint will determine a non-root user, take ownership of necessary directories with this non-root user, and then relaunch the entrypoint as this non-root user.
// When run as non-root, the entrypoint will relaunch itself.
func (e *Entrypoint) bootstrap(ctx context.Context, api Api) error {
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

	_, err = api.RunCommand([]string{executable, "action"}, common.CmdOpts{Attach: true, Env: os.Environ(), User: runAsUser})
	return err
}

// Runs the entrypoint health check.
// Returns an error if the entrypoint health check fails.
func (e *Entrypoint) healthCheck(ctx context.Context, api Api) error {
	if e.HealthCheck == nil {
		return fmt.Errorf("entrypoint health check not defined")
	}

	return e.runCallback(e.HealthCheck)
}

// Prints the version
func (e *Entrypoint) printVersion(ctx context.Context, api Api) error {
	fmt.Print(e.Version)
	return nil
}

// Runs the entrypoint with the provided arguments.
// Returns an error on failure.
func (e *Entrypoint) main(args ...string) error {
	err := e.initialize()
	if err != nil {
		return err
	}

	cmd := "bootstrap"
	if len(args) >= 2 {
		cmd = args[1]
	}

	switch cmd {
	case "action":
		err = e.runCallback(e.action)
	case "bootstrap":
		err = e.runCallback(e.bootstrap)
	case "health":
		err = e.runCallback(e.healthCheck)
	case "version":
		err = e.runCallback(e.printVersion)
	default:
		err = fmt.Errorf("unknown command %s", cmd)
	}

	return err
}

// Runs the entrypoint with the process arguments, and exits on completion.
// Exits with status code 0 on success.
// Exits with status code 1 on failure.
func (e *Entrypoint) Run() {
	err := e.main(os.Args...)

	code := 0
	if err != nil {
		code = 1
		e.Logger.Error("entrypoint failed", "error", err.Error())
	}

	os.Exit(code)
}
