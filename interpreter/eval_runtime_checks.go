package interpreter

import "karl/ast"

func (e *Evaluator) checkRuntimeBeforeEval() error {
	if e.runtime != nil {
		if err := e.runtime.getFatalTaskFailure(); err != nil {
			return err
		}
	}
	if e.currentTask != nil && e.currentTask.canceled() {
		return canceledError()
	}
	return nil
}

func annotateErrorToken(node ast.Node, err error) {
	if err == nil {
		return
	}
	if re, ok := err.(*RuntimeError); ok && re.Token == nil {
		if tok := tokenFromNode(node); tok != nil {
			re.Token = tok
		}
	}
	if re, ok := err.(*RecoverableError); ok && re.Token == nil {
		if tok := tokenFromNode(node); tok != nil {
			re.Token = tok
		}
	}
}

func (e *Evaluator) checkRuntimeAfterEval(sig *Signal, err error) error {
	if err == nil && sig == nil && e.runtime != nil {
		if fatalErr := e.runtime.getFatalTaskFailure(); fatalErr != nil {
			return fatalErr
		}
	}
	return nil
}
