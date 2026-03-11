package interpreter

import (
	"strings"
	"unicode/utf8"
)

type streamIterator interface {
	Next() (Value, bool, error)
	Close() error
}

type streamSourceOpenFunc func(e *Evaluator) (streamIterator, error)
type streamStageApplyFunc func(upstream streamIterator) streamIterator
type streamSinkRunFunc func(e *Evaluator, upstream streamIterator) (Value, error)
type streamSinkPlanRunFunc func(e *Evaluator, plan *StreamPlanValue) (Value, error)

type StreamSourceValue struct {
	name string
	open streamSourceOpenFunc
}

func (s *StreamSourceValue) Type() ValueType { return STREAM_SOURCE }
func (s *StreamSourceValue) Inspect() string { return "<stream-source:" + s.name + ">" }

type StreamStageValue struct {
	name  string
	apply streamStageApplyFunc
}

func (s *StreamStageValue) Type() ValueType { return STREAM_STAGE }
func (s *StreamStageValue) Inspect() string { return "<stream-stage:" + s.name + ">" }

type StreamSinkValue struct {
	name    string
	run     streamSinkRunFunc
	runPlan streamSinkPlanRunFunc
}

func (s *StreamSinkValue) Type() ValueType { return STREAM_SINK }
func (s *StreamSinkValue) Inspect() string { return "<stream-sink:" + s.name + ">" }

type StreamPlanValue struct {
	source *StreamSourceValue
	stages []*StreamStageValue
}

func (s *StreamPlanValue) Type() ValueType { return STREAM_PLAN }
func (s *StreamPlanValue) Inspect() string { return "<stream-plan>" }

func evalStreamPipeInfix(e *Evaluator, left Value, right Value) (Value, error) {
	plan, err := streamPlanFromLeft(left)
	if err != nil {
		return nil, err
	}

	if stage, ok := right.(*StreamStageValue); ok {
		return appendStreamStage(plan, stage), nil
	}
	if sink, ok := right.(*StreamSinkValue); ok {
		return executeStreamPlan(e, plan, sink)
	}
	return nil, &RuntimeError{Message: "operator '|' expects stream stage or sink on the right"}
}

func streamPlanFromLeft(left Value) (*StreamPlanValue, error) {
	switch v := left.(type) {
	case *StreamSourceValue:
		return &StreamPlanValue{source: v, stages: nil}, nil
	case *StreamPlanValue:
		cloned := &StreamPlanValue{source: v.source}
		if len(v.stages) > 0 {
			cloned.stages = append([]*StreamStageValue(nil), v.stages...)
		}
		return cloned, nil
	case *StreamReader:
		return &StreamPlanValue{
			source: &StreamSourceValue{
				name: "stream-reader",
				open: func(_ *Evaluator) (streamIterator, error) {
					return &streamReaderIterator{
						reader:      v,
						size:        defaultStreamReadSize,
						closeOnExit: false,
					}, nil
				},
			},
		}, nil
	case *Process:
		reader, ok := v.outputStream()
		if !ok {
			return nil, recoverableError("process_state", "stdout is only available when stdout mode is \"pipe\"")
		}
		return &StreamPlanValue{
			source: &StreamSourceValue{
				name: "process-stdout",
				open: func(_ *Evaluator) (streamIterator, error) {
					return &streamReaderIterator{
						reader:      reader,
						size:        defaultStreamReadSize,
						closeOnExit: false,
					}, nil
				},
			},
		}, nil
	default:
		return nil, &RuntimeError{Message: "operator '|' expects stream source or plan on the left"}
	}
}

func appendStreamStage(plan *StreamPlanValue, stage *StreamStageValue) *StreamPlanValue {
	if plan == nil {
		return nil
	}
	next := &StreamPlanValue{source: plan.source}
	if len(plan.stages) > 0 {
		next.stages = append([]*StreamStageValue(nil), plan.stages...)
	}
	next.stages = append(next.stages, stage)
	return next
}

func executeStreamPlan(e *Evaluator, plan *StreamPlanValue, sink *StreamSinkValue) (Value, error) {
	if sink == nil {
		return nil, &RuntimeError{Message: "invalid stream sink"}
	}
	if sink.runPlan != nil {
		return sink.runPlan(e, plan)
	}
	if sink.run == nil {
		return nil, &RuntimeError{Message: "invalid stream sink"}
	}

	iter, err := openStreamIterator(e, plan)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = iter.Close()
	}()
	return sink.run(e, iter)
}

func openStreamIterator(e *Evaluator, sourceOrPlan Value) (streamIterator, error) {
	plan, err := streamPlanFromLeft(sourceOrPlan)
	if err != nil {
		return nil, err
	}
	if plan == nil || plan.source == nil || plan.source.open == nil {
		return nil, &RuntimeError{Message: "invalid stream plan source"}
	}
	iter, err := plan.source.open(e)
	if err != nil {
		return nil, err
	}
	for _, stage := range plan.stages {
		if stage == nil || stage.apply == nil {
			_ = iter.Close()
			return nil, &RuntimeError{Message: "invalid stream stage"}
		}
		iter = stage.apply(iter)
	}
	return iter, nil
}

type streamReaderIterator struct {
	reader      *StreamReader
	size        int
	closeOnExit bool
	closed      bool
}

func (s *streamReaderIterator) Next() (Value, bool, error) {
	if s == nil || s.reader == nil {
		return nil, true, nil
	}
	size := s.size
	if size <= 0 {
		size = defaultStreamReadSize
	}
	for {
		chunk, eof, err := s.reader.ReadChunk(size)
		if err != nil {
			return nil, false, err
		}
		if eof {
			return nil, true, nil
		}
		if len(chunk) == 0 {
			continue
		}
		if s.reader.Mode() == streamTypeBytes {
			copied := append([]byte{}, chunk...)
			return &Bytes{Value: copied}, false, nil
		}
		return &String{Value: string(chunk)}, false, nil
	}
}

func (s *streamReaderIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if !s.closeOnExit || s.reader == nil {
		return nil
	}
	return s.reader.Close()
}

type linesIterator struct {
	upstream streamIterator
	pending  []string
	partial  string
	drained  bool
	closed   bool
}

func (l *linesIterator) Next() (Value, bool, error) {
	if l == nil {
		return nil, true, nil
	}
	for {
		if len(l.pending) > 0 {
			out := l.pending[0]
			l.pending = l.pending[1:]
			return &String{Value: out}, false, nil
		}
		if l.drained {
			if l.partial != "" {
				out := l.partial
				l.partial = ""
				return &String{Value: out}, false, nil
			}
			return nil, true, nil
		}

		item, eof, err := l.upstream.Next()
		if err != nil {
			return nil, false, err
		}
		if eof {
			l.drained = true
			continue
		}

		text, err := streamValueToTextChunk(item)
		if err != nil {
			return nil, false, err
		}
		combined := l.partial + text
		parts := strings.Split(combined, "\n")
		if len(parts) == 1 {
			l.partial = combined
			continue
		}
		for _, part := range parts[:len(parts)-1] {
			l.pending = append(l.pending, strings.TrimSuffix(part, "\r"))
		}
		l.partial = parts[len(parts)-1]
	}
}

func (l *linesIterator) Close() error {
	if l == nil || l.closed {
		return nil
	}
	l.closed = true
	if l.upstream != nil {
		return l.upstream.Close()
	}
	return nil
}

func streamValueToTextChunk(v Value) (string, error) {
	switch item := v.(type) {
	case *String:
		return item.Value, nil
	case *Char:
		return item.Value, nil
	case *Bytes:
		return string(item.Value), nil
	default:
		return "", recoverableError("stream_state", "lines() expects bytes or text input")
	}
}

func streamValueToLine(v Value) string {
	switch item := v.(type) {
	case *String:
		return item.Value
	case *Char:
		return item.Value
	case *Bytes:
		if utf8.Valid(item.Value) {
			return string(item.Value)
		}
		return item.Inspect()
	default:
		return formatLogValue(v)
	}
}

func streamValueToBytes(v Value) []byte {
	switch item := v.(type) {
	case *Bytes:
		return item.Value
	case *String:
		return []byte(item.Value)
	case *Char:
		return []byte(item.Value)
	default:
		return []byte(formatLogValue(v))
	}
}
