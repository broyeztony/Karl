package interpreter

import "sync"

const (
	partitionKeyPass = "pass"
	partitionKeyFail = "fail"
)

type streamPartitionMode int

const (
	streamPartitionModeBool streamPartitionMode = iota
	streamPartitionModeKeys
)

func newStreamPartitionSink(e *Evaluator, selector Value, branchKeys []string) *StreamSinkValue {
	mode := streamPartitionModeBool
	keys := []string{partitionKeyPass, partitionKeyFail}
	if len(branchKeys) > 0 {
		mode = streamPartitionModeKeys
		keys = append([]string(nil), branchKeys...)
	}

	return &StreamSinkValue{
		name: "partition",
		runPlan: func(runEval *Evaluator, plan *StreamPlanValue) (Value, error) {
			if runEval == nil {
				runEval = e
			}
			if runEval == nil {
				return nil, &RuntimeError{Message: "partition unavailable"}
			}
			planClone, err := streamPlanFromLeft(plan)
			if err != nil {
				return nil, err
			}
			router := newStreamPartitionRouter(runEval, planClone, selector, mode, keys)
			return router.object(), nil
		},
	}
}

type streamPartitionRouter struct {
	mu sync.Mutex

	eval     *Evaluator
	plan     *StreamPlanValue
	selector Value
	mode     streamPartitionMode

	keys     []string
	branches map[string]*streamPartitionBranch

	started  bool
	upstream streamIterator
	done     bool
	err      error
}

type streamPartitionBranch struct {
	key    string
	opened bool
	closed bool
	queue  []Value
}

func newStreamPartitionRouter(e *Evaluator, plan *StreamPlanValue, selector Value, mode streamPartitionMode, keys []string) *streamPartitionRouter {
	branches := make(map[string]*streamPartitionBranch, len(keys))
	for _, key := range keys {
		branches[key] = &streamPartitionBranch{key: key}
	}
	return &streamPartitionRouter{
		eval:     e,
		plan:     plan,
		selector: selector,
		mode:     mode,
		keys:     append([]string(nil), keys...),
		branches: branches,
	}
}

func (r *streamPartitionRouter) object() *Object {
	pairs := make(map[string]Value, len(r.keys))
	for _, key := range r.keys {
		branchKey := key
		pairs[branchKey] = &StreamSourceValue{
			name: "partition:" + branchKey,
			open: func(_ *Evaluator) (streamIterator, error) {
				return r.openBranch(branchKey)
			},
		}
	}
	return &Object{Pairs: pairs}
}

func (r *streamPartitionRouter) openBranch(key string) (streamIterator, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	branch, ok := r.branches[key]
	if !ok {
		return nil, &RuntimeError{Message: "unknown partition branch: " + key}
	}
	if branch.opened {
		return nil, &RuntimeError{Message: "partition branch already consumed: " + key}
	}
	branch.opened = true
	branch.closed = false

	return &streamPartitionBranchIterator{router: r, key: key}, nil
}

func (r *streamPartitionRouter) nextFor(key string) (Value, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for {
		branch, ok := r.branches[key]
		if !ok || !branch.opened || branch.closed {
			return nil, true, nil
		}
		if len(branch.queue) > 0 {
			next := branch.queue[0]
			branch.queue = branch.queue[1:]
			return next, false, nil
		}
		if r.done {
			if r.err != nil {
				return nil, false, r.err
			}
			return nil, true, nil
		}
		if err := r.ensureUpstreamLocked(); err != nil {
			r.done = true
			r.err = err
			return nil, false, err
		}

		item, eof, err := r.upstream.Next()
		if err != nil {
			r.done = true
			r.err = err
			r.closeUpstreamLocked()
			return nil, false, err
		}
		if eof {
			r.done = true
			r.closeUpstreamLocked()
			return nil, true, nil
		}

		targetKey, err := r.selectBranchKeyLocked(item)
		if err != nil {
			r.done = true
			r.err = err
			r.closeUpstreamLocked()
			return nil, false, err
		}
		target, exists := r.branches[targetKey]
		if !exists || !target.opened || target.closed {
			// Drop unconsumed branches by policy.
			continue
		}

		copied := cloneStreamItem(item)
		if targetKey == key && len(branch.queue) == 0 {
			return copied, false, nil
		}
		target.queue = append(target.queue, copied)
	}
}

func (r *streamPartitionRouter) selectBranchKeyLocked(item Value) (string, error) {
	out, sig, err := r.eval.applyFunction(r.selector, []Value{item})
	if err != nil {
		return "", err
	}
	if sig != nil {
		return "", &RuntimeError{Message: "break/continue outside loop"}
	}

	if r.mode == streamPartitionModeBool {
		b, ok := out.(*Boolean)
		if !ok {
			return "", &RuntimeError{Message: "partition selector must return bool"}
		}
		if b.Value {
			return partitionKeyPass, nil
		}
		return partitionKeyFail, nil
	}

	s, ok := out.(*String)
	if !ok {
		return "", &RuntimeError{Message: "partition selector must return string"}
	}
	return s.Value, nil
}

func (r *streamPartitionRouter) ensureUpstreamLocked() error {
	if r.started {
		return nil
	}
	iter, err := openStreamIterator(r.eval, r.plan)
	if err != nil {
		return err
	}
	r.upstream = iter
	r.started = true
	return nil
}

func (r *streamPartitionRouter) closeUpstreamLocked() {
	if r.upstream == nil {
		return
	}
	_ = r.upstream.Close()
	r.upstream = nil
}

func (r *streamPartitionRouter) closeBranch(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	branch, ok := r.branches[key]
	if !ok || branch.closed {
		return
	}
	branch.closed = true
	branch.queue = nil

	if !r.done && r.started && !r.hasOpenBranchesLocked() {
		r.done = true
		r.closeUpstreamLocked()
	}
}

func (r *streamPartitionRouter) hasOpenBranchesLocked() bool {
	for _, branch := range r.branches {
		if branch.opened && !branch.closed {
			return true
		}
	}
	return false
}

type streamPartitionBranchIterator struct {
	router *streamPartitionRouter
	key    string
	closed bool
}

func (s *streamPartitionBranchIterator) Next() (Value, bool, error) {
	if s == nil || s.closed || s.router == nil {
		return nil, true, nil
	}
	return s.router.nextFor(s.key)
}

func (s *streamPartitionBranchIterator) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.router != nil {
		s.router.closeBranch(s.key)
	}
	return nil
}
