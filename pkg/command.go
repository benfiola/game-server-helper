package helper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// cmdUntilCb is a callback that allows the caller to cancel the command once an external condition has been reached
type cmdUntilCb func(complete func()) error

// a command is an internal extension of [exec.Cmd]
type command struct {
	ctx           context.Context
	ctxCancel     func()
	execCmd       *exec.Cmd
	ignoreSignals bool
	interval      time.Duration
	timeout       time.Duration
	until         cmdUntilCb
}

// Runs the assembled command.
// Returns an error if the command exits with a non-zero exit code.
func (cmd *command) Run() (string, error) {
	Logger(cmd.ctx).Info("run command", "command", cmd.execCmd.Args)

	start := time.Now()
	defer cmd.ctxCancel()

	cmdFinished := make(chan bool, 1)
	var cmdErr error
	stdout := ""
	stderr := ""
	go func() {
		if !cmd.ignoreSignals {
			HandleSignal(cmd.ctx, func(sig os.Signal) {
				cmd.execCmd.Process.Signal(sig)
			})
		}
		cmdErr = cmd.execCmd.Run()
		stdoutBuf, ok := cmd.execCmd.Stdout.(*strings.Builder)
		if ok {
			stdout = stdoutBuf.String()
		}
		stderrBuf, ok := cmd.execCmd.Stderr.(*strings.Builder)
		if ok {
			stderr = stderrBuf.String()
		}
		cmdFinished <- true
	}()

	var cbErr error
	cbFinished := make(chan bool, 1)
	isCmdFinished := false
	go func() {
		if cmd.until == nil {
			cbFinished <- true
			return
		}
		interval := cmd.interval
		if interval == 0 {
			interval = 1 * time.Second
		}
		ticker := time.NewTicker(interval)
		for range ticker.C {
			if isCmdFinished {
				break
			}

			cbErr = cmd.until(cmd.ctxCancel)

			if cbErr != nil {
				cmd.ctxCancel()
				break
			}
		}
		cbFinished <- true
	}()

	isCmdFinished = <-cmdFinished
	<-cbFinished

	select {
	case <-cmd.ctx.Done():
		cmdErr = nil
	default:
	}

	if cmd.timeout > 0 && time.Since(start) > cmd.timeout {
		cmdErr = fmt.Errorf("command timed out")
	}

	err := cbErr
	if err == nil {
		err = cmdErr
		if err != nil {
			Logger(cmd.ctx).Warn("command failed", "cmd", cmd.execCmd.Args, "stderr", stderr, "stdout", stdout)
		}
	}

	return stdout, err
}

// CmdOpts defines the options used in conjunction with the [Command] function
type CmdOpts struct {
	Attach        bool
	Cwd           string
	Env           []string
	IgnoreSignals bool
	Interval      time.Duration
	Until         cmdUntilCb
	User          User
	Timeout       time.Duration
}

// Assembles a command object
func Command(ctx context.Context, cmdSlice []string, opts CmdOpts) *command {
	ctx, ctxCancel := context.WithCancel(ctx)
	if opts.Timeout != 0 {
		ctx, ctxCancel = context.WithTimeout(ctx, opts.Timeout)
	}

	currentUser := GetCurrentUser(ctx)
	if opts.User != (User{}) && opts.User != currentUser {
		cmdSlice = append([]string{"gosu", fmt.Sprintf("%d:%d", opts.User.Uid, opts.User.Gid)}, cmdSlice...)
	}

	stderrBuffer := strings.Builder{}
	stdoutBuffer := strings.Builder{}

	execCmd := exec.CommandContext(ctx, cmdSlice[0], cmdSlice[1:]...)
	execCmd.Stderr = &stderrBuffer
	execCmd.Stdout = &stdoutBuffer
	if opts.Attach {
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
	}
	if opts.Cwd != "" {
		execCmd.Dir = opts.Cwd
	}
	if opts.Env != nil {
		execCmd.Env = opts.Env
	}

	return &command{
		ctx:           ctx,
		ctxCancel:     ctxCancel,
		execCmd:       execCmd,
		ignoreSignals: opts.IgnoreSignals,
		interval:      opts.Interval,
		timeout:       opts.Timeout,
		until:         opts.Until,
	}
}
