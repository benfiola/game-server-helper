package entrypoint

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/benfiola/game-server-helper/pkg/utils"
)

type EntrypointCb func(ctx utils.Context) error

type Entrypoint struct {
	Context       context.Context
	Directories   []string
	EnvRunAsGid   string
	EnvRunAsUid   string
	HealthCb      EntrypointCb
	LocalUsername string
	Logger        *slog.Logger
	MainCb        EntrypointCb
	Version       string
}

func (e Entrypoint) RunWithContext(cb EntrypointCb) error {
	parent := e.Context
	if parent == nil {
		parent = context.Background()
	}
	ctx := utils.Context{Context: parent}
	return cb(ctx)
}

func (e Entrypoint) Bootstrap() error {
	return e.RunWithContext(func(ctx utils.Context) error {
		ctx.Logger().Info("bootstrap")

		runAsUser, err := utils.UserFromCurrent(ctx)
		if err != nil {
			return err
		}

		if runAsUser.Uid == 0 {
			runAsUser, err := utils.UserFromEnv(ctx, e.EnvRunAsUid, e.EnvRunAsGid)
			if err != nil {
				return err
			}

			localUser, err := utils.UserFromUsername(ctx, e.LocalUsername)
			if err != nil {
				return err
			}

			err = localUser.UpdateGidUid(ctx, runAsUser)
			if err != nil {
				return err
			}

			err = utils.SetDirectoryOwner(ctx, runAsUser, e.Directories...)
			if err != nil {
				return err
			}
		}

		executable, err := os.Executable()
		if err != nil {
			return err
		}

		_, err = utils.RunCommand(ctx, []string{executable, "entrypoint"}, utils.CmdOpts{Attach: true, Env: os.Environ(), User: &runAsUser})
		return err
	})
}

func (e Entrypoint) Health() error {
	return e.RunWithContext(e.HealthCb)
}

func (e Entrypoint) Main() error {
	return e.RunWithContext(e.MainCb)
}

func (e Entrypoint) PrintVersion() error {
	return e.RunWithContext(func(ctx utils.Context) error {
		fmt.Print(e.Version)
		return nil
	})
}

func (e Entrypoint) Run() {
	cmd := "bootstrap"
	if len(os.Args) > 0 {
		cmd = os.Args[0]
	}

	var err error
	switch cmd {
	case "bootstrap":
		err = e.Bootstrap()
	case "health":
		err = e.Health()
	case "main":
		err = e.Main()
	case "version":
		err = e.PrintVersion()
	default:
		err = fmt.Errorf("unknown command %s", cmd)
	}

	code := 0
	if err != nil {
		code = 1
		e.Logger.Error("entrypoint failed", "error", err.Error())
	}
	os.Exit(code)
}

type Opts struct {
	Context       context.Context
	Directories   []string
	EnvRunAsGid   string
	EnvRunAsUid   string
	Health        EntrypointCb
	LocalUsername string
	Logger        *slog.Logger
	Main          EntrypointCb
	Version       string
}

func New(opts Opts) (Entrypoint, error) {
	fail := func(err error) (Entrypoint, error) {
		return Entrypoint{}, err
	}

	ctx := opts.Context
	if ctx != nil {
		ctx = context.Background()
	}
	directories := opts.Directories
	if directories == nil {
		directories = []string{}
	}
	envRunAsGid := opts.EnvRunAsGid
	if envRunAsGid == "" {
		envRunAsGid = "GID"
	}
	envRunAsUid := opts.EnvRunAsUid
	if envRunAsUid == "" {
		envRunAsUid = "UID"
	}
	healthCb := opts.Health
	localUsername := opts.LocalUsername
	if localUsername == "" {
		localUsername = "server"
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	}
	mainCb := opts.Main
	if mainCb == nil {
		return fail(fmt.Errorf("field Main is required"))
	}
	version := opts.Version
	if version == "" {
		return fail(fmt.Errorf("field Version is required"))
	}

	return Entrypoint{
		Context:       ctx,
		Directories:   directories,
		EnvRunAsGid:   envRunAsGid,
		EnvRunAsUid:   envRunAsUid,
		HealthCb:      healthCb,
		LocalUsername: localUsername,
		Logger:        logger,
		MainCb:        mainCb,
		Version:       version,
	}, nil
}
