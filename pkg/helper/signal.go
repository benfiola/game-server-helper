package helper

import (
	"os"
	"os/signal"
	"syscall"
)

// SignalHandler holds data required for a signal handler
type SignalHandler struct {
	Api             Api
	HandlerFinished chan bool
	Signal          syscall.Signal
	SignalChannel   chan os.Signal
}

// Unregisters the signal handler
func (sh *SignalHandler) Stop() {
	signal.Stop(sh.SignalChannel)
}

// If a signal is caught, waits for the signal handler to complete.
// If no signal is caught, is a no-op
func (sh *SignalHandler) Wait() {
	if sh.Signal == 0 {
		return
	}

	sh.Api.Logger.Info("wait for signal handler")
	<-sh.HandlerFinished
}

// signalHandlerCb is a callback that is run when a signal is caught
type signalHandlerCb func(sig os.Signal)

// Creates a signal handler and registers it.  When a signal is raised, the provided callback is invoked.
func (api *Api) HandleSignals(cb signalHandlerCb) SignalHandler {
	sh := SignalHandler{
		Api:             *api,
		HandlerFinished: make(chan bool, 1),
		SignalChannel:   make(chan os.Signal, 1),
	}

	signal.Notify(sh.SignalChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if sh.Signal != 0 {
			return
		}

		signal := <-sh.SignalChannel
		api.Logger.Info("signal caught", "signal", signal)
		cb(signal)
		sh.HandlerFinished <- true
	}()

	return sh
}
