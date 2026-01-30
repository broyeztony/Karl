package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"karl/shape"
)

type shapeState struct {
	mu      sync.Mutex
	loaded  map[string]*shape.File
	loading map[string]bool
}

func newShapeState() *shapeState {
	return &shapeState{loaded: make(map[string]*shape.File), loading: make(map[string]bool)}
}

func (e *Evaluator) loadShape(path string) (*shape.File, error) {
	resolved := filepath.Clean(path)
	if abs, err := filepath.Abs(resolved); err == nil {
		resolved = abs
	}

	if e.shapes == nil {
		e.shapes = newShapeState()
	}

	e.shapes.mu.Lock()
	if sh, ok := e.shapes.loaded[resolved]; ok {
		e.shapes.mu.Unlock()
		return sh, nil
	}
	if e.shapes.loading[resolved] {
		e.shapes.mu.Unlock()
		return nil, &RuntimeError{Message: "circular shape import: " + resolved}
	}
	e.shapes.loading[resolved] = true
	e.shapes.mu.Unlock()

	defer func() {
		e.shapes.mu.Lock()
		delete(e.shapes.loading, resolved)
		e.shapes.mu.Unlock()
	}()

	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, &RuntimeError{Message: fmt.Sprintf("shape read error: %v", err)}
	}
	sh, err := shape.ParseFile(string(data))
	if err != nil {
		return nil, &RuntimeError{Message: err.Error()}
	}
	e.shapes.mu.Lock()
	e.shapes.loaded[resolved] = sh
	e.shapes.mu.Unlock()
	return sh, nil
}
