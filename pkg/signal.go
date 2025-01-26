package helper

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// signalHandlerCb is the callback invoked by a signal handler
type signalHandlerCb func(sig os.Signal)

// signalHandlerUnregister is the function that unregisters a registered callback
type signalHandlerUnregister func()

// Allows callers to attach signal handlers to common termination signals to perform cleanup.  Returns an function that unregisters the callback.
func HandleSignal(ctx context.Context, cb signalHandlerCb) signalHandlerUnregister {
	var caught os.Signal
	channel := make(chan os.Signal, 1)

	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		caught = <-channel
		Logger(ctx).Info("signal caught", "signal", caught.String())
		cb(caught)
	}()

	return func() {
		signal.Stop(channel)
	}
}
