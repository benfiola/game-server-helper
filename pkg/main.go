package helper

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
)

// entrypointCb is a callback invoked by the entrypoint.  The callback should receive a context (with which helper methods can be called) and should return an error on failure.
type entrypointCb func(ctx context.Context) error

// An Entrypoint wraps common tasks that need to be performed by many game server docker images.
type Entrypoint struct {
	CheckHealth        entrypointCb
	ctx                context.Context
	Dirs               Map[string, string]
	FileCacheSizeLimit int `env:"CACHE_SIZE_LIMIT"`
	Initialize         func(ctx context.Context) error
	logger             *slog.Logger
	Main               entrypointCb
	uuid               string
	Version            string
}

// 'Bootstraps' the entrypoint.
// When run as root, will determine a non-root user, take ownership of necessary directories with this non-root user, and then relaunch the entrypoint as this non-root user.
// When run as non-root, will directly launch the entrypoint as the non-root user.
func bootstrap(ctx context.Context) error {
	currentUser := GetCurrentUser(ctx)
	runAsUser := currentUser

	if currentUser.Uid == 0 {
		runAsUser, err := GetEnvUser(ctx)
		if err != nil {
			return err
		}

		err = UpdateUser(ctx, "server", runAsUser)
		if err != nil {
			return err
		}

		err = SetOwnerForPaths(ctx, runAsUser, Dirs(ctx).Values()...)
		if err != nil {
			return err
		}
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	_, err = Command(ctx, []string{executable, "entrypoint"}, CmdOpts{Attach: true, Env: os.Environ(), User: runAsUser}).Run()
	return err
}

// Initialies the entrypoint - setting defaults and validating fields.
func (e *Entrypoint) initialize() error {
	err := env.Parse(e)
	if err != nil {
		return err
	}
	e.ctx = context.Background()
	if e.Dirs == nil {
		e.Dirs = Map[string, string]{}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	for key, path := range e.Dirs {
		if filepath.IsAbs(path) {
			continue
		}
		e.Dirs[key] = filepath.Join(wd, path)
	}
	e.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	if e.Main == nil {
		return fmt.Errorf("main unset")
	}
	e.uuid = uuid.NewString()
	if e.Version == "" {
		return fmt.Errorf("version unset")
	}

	e.ctx = context.WithValue(e.ctx, ctxKeyDirs{}, e.Dirs)
	e.ctx = context.WithValue(e.ctx, ctxKeyFileCacheSizeLimit{}, e.FileCacheSizeLimit)
	e.ctx = context.WithValue(e.ctx, ctxKeyLogger{}, e.logger)
	e.ctx = context.WithValue(e.ctx, ctxKeyUuid{}, e.uuid)
	e.ctx = context.WithValue(e.ctx, ctxKeyVersion{}, e.Version)

	if e.Initialize != nil {
		err := e.Initialize(e.ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Prints the version
func printVersion(ctx context.Context) error {
	fmt.Print(Version(ctx))
	return nil
}

// Runs the helper with the provided arguments.
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

	var callback entrypointCb
	switch cmd {
	case "bootstrap":
		callback = bootstrap
	case "entrypoint":
		callback = e.Main
	case "health":
		if e.CheckHealth == nil {
			return fmt.Errorf("check health unimplemented")
		}
		callback = e.CheckHealth
	case "version":
		callback = printVersion
	default:
		return fmt.Errorf("unknown command %s", cmd)
	}

	return callback(e.ctx)
}

// Runs the helper with the process arguments, and exits on completion.
// Exits with status code 0 on success.
// Exits with status code 1 on failure.
func (e *Entrypoint) Run() {
	err := e.main(os.Args...)

	code := 0
	if err != nil {
		code = 1
		e.logger.Error("helper failed", "error", err.Error())
	}

	os.Exit(code)
}
