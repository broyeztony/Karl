//go:build !js

package interpreter

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func registerSignalBuiltins() {
	builtins["signalWatch"] = &Builtin{Name: "signalWatch", Fn: builtinSignalWatch}
}

func builtinSignalWatch(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "signalWatch expects 1 argument"}
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "signalWatch expects array of signal names"}
	}
	if len(arr.Elements) == 0 {
		return nil, &RuntimeError{Message: "signalWatch expects at least one signal"}
	}

	watch := make([]os.Signal, 0, len(arr.Elements))
	names := make(map[os.Signal]string, len(arr.Elements))
	for _, el := range arr.Elements {
		name, ok := stringArg(el)
		if !ok {
			return nil, &RuntimeError{Message: "signalWatch expects string signal names"}
		}
		sig, canonical, ok := signalFromName(name)
		if !ok {
			return nil, &RuntimeError{Message: "signalWatch unknown signal: " + name}
		}
		watch = append(watch, sig)
		names[sig] = canonical
	}

	out := &Channel{Ch: make(chan Value, len(watch)+4)}
	osCh := make(chan os.Signal, len(watch)+4)
	watchDone := make(chan struct{})
	signal.Notify(osCh, watch...)
	out.onClose = func() {
		signal.Stop(osCh)
		close(watchDone)
	}

	cancelCh := runtimeCancelSignal(e)
	fatalCh := runtimeFatalSignal(e)
	go func() {
		defer out.Close()
		for {
			select {
			case sig := <-osCh:
				if sig == nil {
					continue
				}
				name, ok := names[sig]
				if !ok {
					name = strings.ToUpper(sig.String())
				}
				if !channelTrySend(out, &String{Value: name}) {
					return
				}
			case <-cancelCh:
				return
			case <-fatalCh:
				return
			case <-watchDone:
				return
			}
		}
	}()

	return out, nil
}

func signalFromName(name string) (os.Signal, string, bool) {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "SIGINT":
		return os.Interrupt, "SIGINT", true
	case "SIGTERM":
		return syscall.SIGTERM, "SIGTERM", true
	case "SIGHUP":
		return syscall.SIGHUP, "SIGHUP", true
	case "SIGQUIT":
		return syscall.SIGQUIT, "SIGQUIT", true
	case "SIGUSR1":
		return syscall.SIGUSR1, "SIGUSR1", true
	case "SIGUSR2":
		return syscall.SIGUSR2, "SIGUSR2", true
	default:
		return nil, "", false
	}
}

func channelTrySend(ch *Channel, val Value) (ok bool) {
	if ch == nil || ch.Closed {
		return false
	}
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	ch.Ch <- val
	return true
}
