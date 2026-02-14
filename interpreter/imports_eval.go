package interpreter

import (
	"fmt"

	"karl/ast"
)

func (e *Evaluator) evalImportExpression(node *ast.ImportExpression, _ *Environment) (Value, *Signal, error) {
	if node.Path == nil {
		return nil, nil, &RuntimeError{Message: "import expects a string literal path"}
	}
	path, err := e.resolveImportPath(node.Path.Value)
	if err != nil {
		return nil, nil, &RuntimeError{Message: "import path error: " + err.Error()}
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
				runtime:     e.runtime,
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
