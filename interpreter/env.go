package interpreter

import "sync"

type Environment struct {
	mu    sync.RWMutex
	store map[string]Value
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Value)}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Value, bool) {
	e.mu.RLock()
	val, ok := e.store[name]
	e.mu.RUnlock()
	if ok {
		return val, true
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}

func (e *Environment) GetLocal(name string) (Value, bool) {
	e.mu.RLock()
	val, ok := e.store[name]
	e.mu.RUnlock()
	return val, ok
}

func (e *Environment) Define(name string, val Value) {
	e.mu.Lock()
	e.store[name] = val
	e.mu.Unlock()
}

func (e *Environment) Set(name string, val Value) bool {
	e.mu.Lock()
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		e.mu.Unlock()
		return true
	}
	e.mu.Unlock()
	if e.outer != nil {
		return e.outer.Set(name, val)
	}
	return false
}

func (e *Environment) Snapshot() map[string]Value {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make(map[string]Value, len(e.store))
	for k, v := range e.store {
		out[k] = v
	}
	return out
}

// SnapshotAll returns a merged snapshot of the current scope and all outers.
// Inner scope bindings shadow outer ones.
func (e *Environment) SnapshotAll() map[string]Value {
	if e == nil {
		return map[string]Value{}
	}
	local := e.Snapshot()
	if e.outer == nil {
		return local
	}
	out := e.outer.SnapshotAll()
	for k, v := range local {
		out[k] = v
	}
	return out
}
