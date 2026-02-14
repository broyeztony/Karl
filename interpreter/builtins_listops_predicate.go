package interpreter

func applyListPredicate(e *Evaluator, fn Value, el Value, op string) (bool, error) {
	val, _, err := e.applyFunction(fn, []Value{el})
	if err != nil {
		return false, err
	}
	b, ok := val.(*Boolean)
	if !ok {
		return false, &RuntimeError{Message: op + " predicate must return bool"}
	}
	return b.Value, nil
}
