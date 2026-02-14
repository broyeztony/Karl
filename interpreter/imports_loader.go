package interpreter

import (
	"fmt"
	"os"
	"path/filepath"

	"karl/ast"
	"karl/lexer"
	"karl/parser"
)

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
