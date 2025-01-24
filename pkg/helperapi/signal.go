package helperapi

import (
	"os"
	"os/signal"
	"syscall"
)

// SignalHandlerCb is the callback invoked by a signal handler
type SignalHandlerCb func(sig os.Signal)

// SignalHandlerUnregister is the function that unregisters a registered callback
type SignalHandlerUnregister func()

// Allows callers to attach signal handlers to common termination signals to perform cleanup.  Returns an function that unregisters the callback.
func (api *Api) HandleSignal(cb SignalHandlerCb) SignalHandlerUnregister {
	var caught os.Signal
	channel := make(chan os.Signal, 1)

	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		caught = <-channel
		cb(caught)
	}()

	return func() {
		signal.Stop(channel)
	}
}
