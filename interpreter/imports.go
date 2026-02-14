package interpreter

import (
	"sync"

	"karl/ast"
)

type moduleState struct {
	mu      sync.Mutex
	loaded  map[string]*moduleDefinition
	loading map[string]bool
}

func newModuleState() *moduleState {
	return &moduleState{
		loaded:  make(map[string]*moduleDefinition),
		loading: make(map[string]bool),
	}
}

type moduleDefinition struct {
	program  *ast.Program
	source   string
	filename string
}
