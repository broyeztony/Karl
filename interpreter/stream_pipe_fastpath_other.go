//go:build !js && !linux

package interpreter

import "io"

func streamPipeTryFastCopy(_ io.Reader, _ io.Writer, _ int) (int64, int64, bool, error) {
	return 0, 0, false, nil
}
