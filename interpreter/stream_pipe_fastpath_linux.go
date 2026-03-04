//go:build !js && linux

package interpreter

import (
	"errors"
	"io"

	"golang.org/x/sys/unix"
)

func streamPipeTryFastCopy(src io.Reader, dst io.Writer, bufferSize int) (int64, int64, bool, error) {
	srcFD, ok := streamFD(src)
	if !ok {
		return 0, 0, false, nil
	}
	dstFD, ok := streamFD(dst)
	if !ok {
		return 0, 0, false, nil
	}
	if bufferSize <= 0 {
		bufferSize = defaultPipeBufferSize
	}

	srcPipe, srcRegular, ok := streamFDKinds(srcFD)
	if !ok {
		return 0, 0, false, nil
	}
	dstPipe, _, ok := streamFDKinds(dstFD)
	if !ok {
		return 0, 0, false, nil
	}

	// Prefer splice whenever one end is a pipe.
	if srcPipe || dstPipe {
		total, chunks, err := spliceLoop(srcFD, dstFD, bufferSize)
		if err != nil && total == 0 && isUnsupportedFastPathErr(err) {
			return 0, 0, false, nil
		}
		return total, chunks, true, err
	}

	// Then try sendfile for regular-file sources.
	if srcRegular {
		total, chunks, err := sendfileLoop(srcFD, dstFD, bufferSize)
		if err != nil && total == 0 && isUnsupportedFastPathErr(err) {
			return 0, 0, false, nil
		}
		return total, chunks, true, err
	}

	return 0, 0, false, nil
}

func streamFDKinds(fd int) (isPipe bool, isRegular bool, ok bool) {
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		return false, false, false
	}
	mode := st.Mode & unix.S_IFMT
	return mode == unix.S_IFIFO, mode == unix.S_IFREG, true
}

func streamFD(v interface{}) (int, bool) {
	type hasFD interface {
		Fd() uintptr
	}
	fdCarrier, ok := v.(hasFD)
	if !ok {
		return 0, false
	}
	fd := int(fdCarrier.Fd())
	if fd < 0 {
		return 0, false
	}
	return fd, true
}

func spliceLoop(srcFD int, dstFD int, chunkSize int) (int64, int64, error) {
	var total int64
	var chunks int64
	for {
		n, err := unix.Splice(srcFD, nil, dstFD, nil, chunkSize, 0)
		if n > 0 {
			total += int64(n)
			chunks++
			continue
		}
		if err == nil {
			// EOF on source.
			return total, chunks, nil
		}
		if errors.Is(err, unix.EINTR) || errors.Is(err, unix.EAGAIN) {
			continue
		}
		return total, chunks, err
	}
}

func sendfileLoop(srcFD int, dstFD int, chunkSize int) (int64, int64, error) {
	var total int64
	var chunks int64
	for {
		n, err := unix.Sendfile(dstFD, srcFD, nil, chunkSize)
		if n > 0 {
			total += int64(n)
			chunks++
			continue
		}
		if err == nil {
			// EOF on source.
			return total, chunks, nil
		}
		if errors.Is(err, unix.EINTR) || errors.Is(err, unix.EAGAIN) {
			continue
		}
		return total, chunks, err
	}
}

func isUnsupportedFastPathErr(err error) bool {
	return errors.Is(err, unix.ENOSYS) ||
		errors.Is(err, unix.EINVAL) ||
		errors.Is(err, unix.ENOTSUP) ||
		errors.Is(err, unix.EOPNOTSUPP) ||
		errors.Is(err, unix.EXDEV) ||
		errors.Is(err, unix.EPERM) ||
		errors.Is(err, unix.ENOTSOCK)
}
