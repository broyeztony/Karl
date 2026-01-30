package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"karl/ast"
	"karl/lexer"
	"karl/parser"
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

func (e *Evaluator) evalImportExpression(node *ast.ImportExpression, _ *Environment) (Value, *Signal, error) {
	if node.Path == nil {
		return nil, nil, &RuntimeError{Message: "import expects a string literal path"}
	}
	path, err := e.resolveImportPath(node.Path.Value)
	if err != nil {
		return nil, nil, &RuntimeError{Message: "import path error: " + err.Error()}
	}
	if strings.HasSuffix(path, ".shape") {
		shFile, err := e.loadShape(path)
		if err != nil {
			return nil, nil, err
		}
		if len(shFile.Shapes) == 1 {
			sh := shFile.Shapes[0]
			return &ShapeValue{Name: sh.Name, Shape: sh.Type}, nil, nil
		}
		pairs := make(map[string]Value, len(shFile.Shapes))
		for _, sh := range shFile.Shapes {
			pairs[sh.Name] = &ShapeValue{Name: sh.Name, Shape: sh.Type}
		}
		return &Object{Pairs: pairs}, nil, nil
	}
	module, err := e.loadModule(path)
	if err != nil {
		return nil, nil, err
	}
	factory := &Builtin{
		Name: "moduleFactory",
		Fn: func(_ *Evaluator, args []Value) (Value, error) {
			if len(args) != 0 {
				return nil, &RuntimeError{Message: "module factory expects no arguments"}
			}
			moduleEnv := NewEnclosedEnvironment(NewBaseEnvironment())
			moduleEval := &Evaluator{
				source:      module.source,
				filename:    module.filename,
				projectRoot: e.projectRoot,
				modules:     e.modules,
			}
			val, sig, err := moduleEval.Eval(module.program, moduleEnv)
			if err != nil {
				return nil, fmt.Errorf("%s", FormatRuntimeError(err, moduleEval.source, moduleEval.filename))
			}
			if sig != nil {
				return nil, &RuntimeError{Message: "break/continue outside loop"}
			}
			_ = val
			return &ModuleObject{Env: moduleEnv}, nil
		},
	}
	return factory, nil, nil
}

func (e *Evaluator) resolveImportPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	root := e.projectRoot
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		root = cwd
	}
	return filepath.Join(root, path), nil
}

func (e *Evaluator) loadModule(path string) (*moduleDefinition, error) {
	resolved := filepath.Clean(path)
	if abs, err := filepath.Abs(resolved); err == nil {
		resolved = abs
	}

	if e.modules == nil {
		e.modules = newModuleState()
	}

	e.modules.mu.Lock()
	if module, ok := e.modules.loaded[resolved]; ok {
		e.modules.mu.Unlock()
		return module, nil
	}
	if e.modules.loading[resolved] {
		e.modules.mu.Unlock()
		return nil, &RuntimeError{Message: "circular import: " + resolved}
	}
	e.modules.loading[resolved] = true
	e.modules.mu.Unlock()

	defer func() {
		e.modules.mu.Lock()
		delete(e.modules.loading, resolved)
		e.modules.mu.Unlock()
	}()

	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, &RuntimeError{Message: fmt.Sprintf("import read error: %v", err)}
	}
	program, err := parseImportProgram(data, resolved)
	if err != nil {
		return nil, err
	}

	definition := &moduleDefinition{
		program:  program,
		source:   string(data),
		filename: resolved,
	}
	e.modules.mu.Lock()
	e.modules.loaded[resolved] = definition
	e.modules.mu.Unlock()
	return definition, nil
}

func parseImportProgram(data []byte, filename string) (*ast.Program, error) {
	p := parser.New(lexer.New(string(data)))
	program := p.ParseProgram()
	if errs := p.ErrorsDetailed(); len(errs) > 0 {
		return nil, fmt.Errorf("%s", parser.FormatParseErrors(errs, string(data), filename))
	}
	return program, nil
}
