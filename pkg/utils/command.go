package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CmdOpts struct {
	Attach bool
	Cwd    string
	Debug  bool
	Env    []string
	User   *User
}

func RunCommand(ctx Context, cmdSlice []string, opts CmdOpts) (string, error) {
	fail := func(err error) (string, error) {
		return "", err
	}

	currentUser, err := UserFromCurrent(ctx)
	if err != nil {
		return fail(err)
	}

	if opts.User != nil && *opts.User != currentUser {
		cmdSlice = append([]string{"gosu", fmt.Sprintf("%d:%d", opts.User.Uid, opts.User.Gid)}, cmdSlice...)
	}

	ctx.Logger().Info("run cmd", "command", cmdSlice)

	stderrBuffer := strings.Builder{}
	stdoutBuffer := strings.Builder{}
	command := exec.CommandContext(ctx, cmdSlice[0], cmdSlice[1:]...)
	command.Stderr = &stderrBuffer
	command.Stdout = &stdoutBuffer
	if opts.Attach {
		command.Stderr = os.Stderr
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
	}
	if opts.Cwd != "" {
		command.Dir = opts.Cwd
	}
	if opts.Env != nil {
		command.Env = opts.Env
	}

	err = command.Run()
	if err != nil && opts.Debug {
		ctx.Logger().Error("run cmd failed", "command", cmdSlice, "stdout", stdoutBuffer.String(), "stderr", stderrBuffer.String())
	}

	return stdoutBuffer.String(), err
}
