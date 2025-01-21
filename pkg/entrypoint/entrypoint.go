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
	HealthCb      EntrypointCb
	LocalUsername string
	Logger        *slog.Logger
	MainCb        EntrypointCb
	RunAsUser     utils.User
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
			runAsUser = e.RunAsUser

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
	Health        EntrypointCb
	LocalUsername string
	Logger        *slog.Logger
	Main          EntrypointCb
	RunAsUser     utils.User
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
	healthCb := opts.Health
	localUsername := opts.LocalUsername
	if localUsername == "" {
		return fail(fmt.Errorf("field LocalUsername is required"))
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	}
	mainCb := opts.Main
	if mainCb == nil {
		return fail(fmt.Errorf("field Main is required"))
	}
	runAsUser := opts.RunAsUser
	if runAsUser == (utils.User{}) {
		return fail(fmt.Errorf("field RunAsUser is required"))
	}
	version := opts.Version
	if version == "" {
		return fail(fmt.Errorf("field Version is required"))
	}

	return Entrypoint{
		Context:       ctx,
		Directories:   directories,
		HealthCb:      healthCb,
		LocalUsername: localUsername,
		Logger:        logger,
		MainCb:        mainCb,
		RunAsUser:     runAsUser,
		Version:       version,
	}, nil
}
