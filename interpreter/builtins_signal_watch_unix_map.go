//go:build !js && !windows

package interpreter

import (
	"os"
	"syscall"
)

func platformSignalFromName(name string) (os.Signal, string, bool) {
	switch name {
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
