package interpreter

import "fmt"

const (
	TaskFailurePolicyFailFast = "fail-fast"
	TaskFailurePolicyDefer    = "defer"
)

func validTaskFailurePolicy(policy string) bool {
	return policy == TaskFailurePolicyFailFast || policy == TaskFailurePolicyDefer
}

func (r *runtimeState) setTaskFailurePolicy(policy string) error {
	if r == nil {
		return nil
	}
	if !validTaskFailurePolicy(policy) {
		return fmt.Errorf("invalid task failure policy: %s", policy)
	}
	r.mu.Lock()
	r.taskFailurePolicy = policy
	r.mu.Unlock()
	return nil
}

func (r *runtimeState) getTaskFailurePolicy() string {
	if r == nil {
		return TaskFailurePolicyFailFast
	}
	r.mu.Lock()
	policy := r.taskFailurePolicy
	r.mu.Unlock()
	if policy == "" {
		return TaskFailurePolicyFailFast
	}
	return policy
}

func (r *runtimeState) setFatalTaskFailure(err error) {
	if r == nil || err == nil {
		return
	}
	shouldSignal := false
	r.mu.Lock()
	if r.fatalTaskFailure == nil {
		r.fatalTaskFailure = err
		shouldSignal = true
	}
	r.mu.Unlock()
	if shouldSignal {
		r.fatalOnce.Do(func() {
			close(r.fatalCh)
		})
	}
}

func (r *runtimeState) getFatalTaskFailure() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	err := r.fatalTaskFailure
	r.mu.Unlock()
	return err
}

func (r *runtimeState) fatalSignal() <-chan struct{} {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	ch := r.fatalCh
	r.mu.Unlock()
	return ch
}
