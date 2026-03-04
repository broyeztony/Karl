package interpreter

import (
	"errors"
	"io"
	"sync"
)

const (
	streamTypeText  = "text"
	streamTypeBytes = "bytes"

	defaultStreamReadSize = 4096
	defaultPipeBufferSize = 32768
)

type StreamReader struct {
	mu sync.Mutex

	reader io.Reader
	closer io.Closer

	closed bool
	mode   string
}

func (s *StreamReader) Type() ValueType { return STREAM_READER }
func (s *StreamReader) Inspect() string { return "<stream-reader>" }

func (s *StreamReader) Mode() string {
	if s == nil {
		return streamTypeText
	}
	s.mu.Lock()
	mode := s.mode
	s.mu.Unlock()
	if mode == "" {
		return streamTypeText
	}
	return mode
}

func (s *StreamReader) ReadChunk(size int) ([]byte, bool, error) {
	if s == nil {
		return nil, false, errors.New("stream reader unavailable")
	}
	if size <= 0 {
		size = defaultStreamReadSize
	}

	buf := make([]byte, size)

	s.mu.Lock()
	closed := s.closed
	reader := s.reader
	s.mu.Unlock()

	if closed {
		return nil, true, nil
	}
	if reader == nil {
		return nil, true, nil
	}

	n, err := reader.Read(buf)
	if n > 0 {
		return buf[:n], false, nil
	}
	if errors.Is(err, io.EOF) {
		return nil, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	return nil, false, nil
}

func (s *StreamReader) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	closer := s.closer
	s.mu.Unlock()

	if closer != nil {
		return closer.Close()
	}
	return nil
}

type StreamWriter struct {
	mu sync.Mutex

	writer io.Writer
	closer io.Closer

	closed bool
	mode   string
}

func (s *StreamWriter) Type() ValueType { return STREAM_WRITER }
func (s *StreamWriter) Inspect() string { return "<stream-writer>" }

func (s *StreamWriter) Mode() string {
	if s == nil {
		return streamTypeText
	}
	s.mu.Lock()
	mode := s.mode
	s.mu.Unlock()
	if mode == "" {
		return streamTypeText
	}
	return mode
}

func (s *StreamWriter) WriteChunk(data []byte) (int, error) {
	if s == nil {
		return 0, errors.New("stream writer unavailable")
	}

	s.mu.Lock()
	closed := s.closed
	writer := s.writer
	s.mu.Unlock()

	if closed {
		return 0, errors.New("stream writer is closed")
	}
	if writer == nil {
		return 0, errors.New("stream writer unavailable")
	}
	return writer.Write(data)
}

func (s *StreamWriter) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	closer := s.closer
	s.mu.Unlock()

	if closer != nil {
		return closer.Close()
	}
	return nil
}
