package helper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CmdOpts defines the options used in conjunction with the [RunCommand] function
type CmdOpts struct {
	Attach        bool
	Context       context.Context
	Cwd           string
	Env           []string
	IgnoreSignals bool
	User          User
}

// Runs a command (defined as a string slice) and returns the stdout.
// Raises an error if the command fails.
func (api *Api) RunCommand(cmdSlice []string, opts CmdOpts) (string, error) {
	fail := func(err error) (string, error) {
		return "", err
	}

	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	currentUser, err := api.GetCurrentUser()
	if err != nil {
		return fail(err)
	}
	if opts.User != (User{}) && opts.User != currentUser {
		cmdSlice = append([]string{"gosu", fmt.Sprintf("%d:%d", opts.User.Uid, opts.User.Gid)}, cmdSlice...)
	}

	api.Logger.Info("run cmd", "command", cmdSlice)

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
		api.HandleSignals(func(sig os.Signal) {
			command.Process.Signal(sig)
		})
	}

	err = command.Run()
	if err != nil && !opts.Attach {
		api.Logger.Error("run cmd failed", "command", cmdSlice, "stderr", stderrBuffer.String())
	}

	return stdoutBuffer.String(), err
}

// cmdUntilCb is a callback that polls for a condition.  Once this condition is reached, the completion callback should be called.
// Return an error to fail the [RunCommandUntil] function
type cmdUntilCb func(complete func()) error

// CmdUntilOpts defines the options used in conjunction with the [RunCommandUntil] function
type CmdUntilOpts struct {
	CmdOpts
	Callback cmdUntilCb
	Interval time.Duration
	Timeout  time.Duration
}

// Runs a command until a condition (dictated by [CmdUntilOpts.Callback]) or a timeout is reached.
// Returns a failure if the command fails.
// Returns a failure if the callback fails.
// Returns a failure if a timeout is reached.
func (api *Api) RunCommandUntil(cmdSlice []string, opts CmdUntilOpts) error {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	interval := opts.Interval
	if interval == 0 {
		interval = 1 * time.Second
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 120 * time.Second
	}

	start := time.Now()
	sctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmdFinished := make(chan bool, 1)
	var cmdErr error
	go func() {
		_, cmdErr = api.RunCommand(cmdSlice, CmdOpts{
			Attach:        opts.Attach,
			Context:       sctx,
			Cwd:           opts.Cwd,
			Env:           opts.Env,
			IgnoreSignals: opts.IgnoreSignals,
			User:          opts.User,
		})
		cmdFinished <- true
	}()

	var cbErr error
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			isCmdFinished := false
			select {
			case <-cmdFinished:
				isCmdFinished = true
			default:
			}

			if isCmdFinished {
				break
			}

			cbErr = opts.Callback(cancel)

			if cbErr != nil {
				cancel()
				break
			}
		}
	}()

	<-cmdFinished

	select {
	case <-sctx.Done():
		cmdErr = nil
	default:
	}

	if time.Since(start) > timeout {
		cmdErr = fmt.Errorf("command timed out")
	}

	err := cbErr
	if err == nil {
		err = cmdErr
	}

	return err
}
