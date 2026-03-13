package interpreter

import (
	"fmt"
	"time"
)

const runtimeBlockProbeInterval = 25 * time.Millisecond

func panicToError(context string, recovered interface{}) error {
	msg := fmt.Sprintf("%v", recovered)
	if context == "" {
		return &RuntimeError{Message: "runtime panic: " + msg}
	}
	return &RuntimeError{Message: context + " panic: " + msg}
}

func reportRuntimePanic(runtime *runtimeState, context string, recovered interface{}) {
	if recovered == nil || runtime == nil {
		return
	}
	runtime.setFatalTaskFailure(panicToError(context, recovered))
}

func runGuarded(runtime *runtimeState, context string, fn func()) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				reportRuntimePanic(runtime, context, recovered)
			}
		}()
		fn()
	}()
}
