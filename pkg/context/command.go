package context

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CmdOpts defines the options used in conjunction with the [RunCommand] function
type CmdOpts struct {
	Attach        bool
	Cwd           string
	Env           []string
	IgnoreSignals bool
	User          User
}

// Runs a command (defined as a string slice) and returns the stdout.
// Raises an error if the command fails.
func (ctx *Context) RunCommand(cmdSlice []string, opts CmdOpts) (string, error) {
	fail := func(err error) (string, error) {
		return "", err
	}

	currentUser, err := ctx.GetCurrentUser()
	if err != nil {
		return fail(err)
	}
	if opts.User != (User{}) && opts.User != currentUser {
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
	if !opts.IgnoreSignals {
		ctx.HandleSignals(func(sig os.Signal) {
			command.Process.Signal(sig)
		})
	}

	err = command.Run()
	if err != nil && !opts.Attach {
		ctx.Logger().Error("run cmd failed", "command", cmdSlice, "stderr", stderrBuffer.String())
	}

	return stdoutBuffer.String(), err
}
