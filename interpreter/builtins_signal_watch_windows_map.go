//go:build windows

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
	default:
		return nil, "", false
	}
}
