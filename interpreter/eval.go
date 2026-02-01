package interpreter

import (
	"fmt"
	"karl/ast"
	"sort"
	"unicode/utf8"
)

type Evaluator struct {
	source      string
	filename    string
	projectRoot string
	modules     *moduleState
}

func NewEvaluator() *Evaluator {
	return &Evaluator{modules: newModuleState()}
}

func NewEvaluatorWithSource(source string) *Evaluator {
	return &Evaluator{source: source, modules: newModuleState()}
}

func NewEvaluatorWithSourceAndFilename(source string, filename string) *Evaluator {
	return &Evaluator{source: source, filename: filename, modules: newModuleState()}
}

func NewEvaluatorWithSourceFilenameAndRoot(source string, filename string, root string) *Evaluator {
	return &Evaluator{
		source:      source,
		filename:    filename,
		projectRoot: root,
		modules:     newModuleState(),
	}
}

func (e *Evaluator) SetProjectRoot(root string) {
	e.projectRoot = root
}

func (e *Evaluator) formatError(err error) string {
	return FormatRuntimeError(err, e.source, e.filename)
}

type queryRow struct {
	item Value
	key  Value
}

func recoverableErrorValue(err *RecoverableError) Value {
	kind := err.Kind
	if kind == "" {
		kind = "error"
	}
	return &Object{Pairs: map[string]Value{
		"kind":    &String{Value: kind},
		"message": &String{Value: err.Message},
	}}
}

func (e *Evaluator) Eval(node ast.Node, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.evalNode(node, env)
	if err != nil {
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
	return val, sig, err
}

func (e *Evaluator) evalNode(node ast.Node, env *Environment) (Value, *Signal, error) {
	switch n := node.(type) {
	case *ast.Program:
		return e.evalProgram(n, env)
	case *ast.ExpressionStatement:
		return e.Eval(n.Expression, env)
	case *ast.LetStatement:
		val, sig, err := e.Eval(n.Value, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		if ok, err := bindPattern(n.Name, val, env); !ok || err != nil {
			if err != nil {
				return nil, nil, err
			}
			return nil, nil, &RuntimeError{Message: "let pattern did not match"}
		}
		return UnitValue, nil, nil
	case *ast.Identifier:
		return e.evalIdentifier(n, env)
	case *ast.Placeholder:
		return nil, nil, &RuntimeError{Message: "placeholder is only valid in call arguments"}
	case *ast.IntegerLiteral:
		return &Integer{Value: n.Value}, nil, nil
	case *ast.FloatLiteral:
		return &Float{Value: n.Value}, nil, nil
	case *ast.StringLiteral:
		return &String{Value: n.Value}, nil, nil
	case *ast.CharLiteral:
		return &Char{Value: n.Value}, nil, nil
	case *ast.BooleanLiteral:
		return &Boolean{Value: n.Value}, nil, nil
	case *ast.NullLiteral:
		return NullValue, nil, nil
	case *ast.UnitLiteral:
		return UnitValue, nil, nil
	case *ast.PrefixExpression:
		return e.evalPrefixExpression(n, env)
	case *ast.InfixExpression:
		return e.evalInfixExpression(n, env)
	case *ast.AssignExpression:
		return e.evalAssignExpression(n, env)
	case *ast.PostfixExpression:
		return e.evalPostfixExpression(n, env)
	case *ast.AwaitExpression:
		return e.evalAwaitExpression(n, env)
	case *ast.ImportExpression:
		return e.evalImportExpression(n, env)
	case *ast.RecoverExpression:
		return e.evalRecoverExpression(n, env)
	case *ast.IfExpression:
		return e.evalIfExpression(n, env)
	case *ast.BlockExpression:
		return e.evalBlockExpression(n, env)
	case *ast.MatchExpression:
		return e.evalMatchExpression(n, env)
	case *ast.ForExpression:
		return e.evalForExpression(n, env)
	case *ast.LambdaExpression:
		return &Function{Params: n.Params, Body: n.Body, Env: env}, nil, nil
	case *ast.CallExpression:
		return e.evalCallExpression(n, env)
	case *ast.MemberExpression:
		return e.evalMemberExpression(n, env)
	case *ast.IndexExpression:
		return e.evalIndexExpression(n, env)
	case *ast.SliceExpression:
		return e.evalSliceExpression(n, env)
	case *ast.ArrayLiteral:
		return e.evalArrayLiteral(n, env)
	case *ast.ObjectLiteral:
		return e.evalObjectLiteral(n, env)
	case *ast.StructInitExpression:
		return e.evalStructInitExpression(n, env)
	case *ast.RangeExpression:
		return e.evalRangeExpression(n, env)
	case *ast.QueryExpression:
		return e.evalQueryExpression(n, env)
	case *ast.RaceExpression:
		return e.evalRaceExpression(n, env)
	case *ast.SpawnExpression:
		return e.evalSpawnExpression(n, env)
	case *ast.BreakExpression:
		return e.evalBreakExpression(n, env)
	case *ast.ContinueExpression:
		return UnitValue, &Signal{Type: SignalContinue}, nil
	default:
		return nil, nil, &RuntimeError{Message: fmt.Sprintf("unsupported node: %T", node)}
	}
}

func (e *Evaluator) evalProgram(program *ast.Program, env *Environment) (Value, *Signal, error) {
	var result Value = UnitValue
	for _, stmt := range program.Statements {
		val, sig, err := e.Eval(stmt, env)
		if err != nil {
			return nil, nil, err
		}
		if sig != nil {
			return nil, nil, &RuntimeError{Message: "break/continue outside loop"}
		}
		result = val
	}
	return result, nil, nil
}

func (e *Evaluator) evalIdentifier(node *ast.Identifier, env *Environment) (Value, *Signal, error) {
	val, ok := env.Get(node.Value)
	if !ok {
		return nil, nil, &RuntimeError{Message: "undefined identifier: " + node.Value}
	}
	return val, nil, nil
}

func (e *Evaluator) evalPrefixExpression(node *ast.PrefixExpression, env *Environment) (Value, *Signal, error) {
	right, sig, err := e.Eval(node.Right, env)
	if err != nil || sig != nil {
		return right, sig, err
	}
	switch node.Operator {
	case "!":
		b, ok := right.(*Boolean)
		if !ok {
			return nil, nil, &RuntimeError{Message: "operator ! expects bool"}
		}
		return &Boolean{Value: !b.Value}, nil, nil
	case "-":
		switch v := right.(type) {
		case *Integer:
			return &Integer{Value: -v.Value}, nil, nil
		case *Float:
			return &Float{Value: -v.Value}, nil, nil
		default:
			return nil, nil, &RuntimeError{Message: "operator - expects number"}
		}
	default:
		return nil, nil, &RuntimeError{Message: "unknown prefix operator: " + node.Operator}
	}
}

func (e *Evaluator) evalInfixExpression(node *ast.InfixExpression, env *Environment) (Value, *Signal, error) {
	left, sig, err := e.Eval(node.Left, env)
	if err != nil || sig != nil {
		return left, sig, err
	}

	if node.Operator == "&&" || node.Operator == "||" {
		lb, ok := left.(*Boolean)
		if !ok {
			return nil, nil, &RuntimeError{Message: "logical operators require bool"}
		}
		if node.Operator == "&&" && !lb.Value {
			return &Boolean{Value: false}, nil, nil
		}
		if node.Operator == "||" && lb.Value {
			return &Boolean{Value: true}, nil, nil
		}
		right, sig, err := e.Eval(node.Right, env)
		if err != nil || sig != nil {
			return right, sig, err
		}
		rb, ok := right.(*Boolean)
		if !ok {
			return nil, nil, &RuntimeError{Message: "logical operators require bool"}
		}
		return &Boolean{Value: rb.Value}, nil, nil
	}

	right, sig, err := e.Eval(node.Right, env)
	if err != nil || sig != nil {
		return right, sig, err
	}

	switch node.Operator {
	case "==":
		return &Boolean{Value: StrictEqual(left, right)}, nil, nil
	case "!=":
		return &Boolean{Value: !StrictEqual(left, right)}, nil, nil
	case "eqv":
		return &Boolean{Value: Equivalent(left, right)}, nil, nil
	}

	switch l := left.(type) {
	case *Integer:
		return evalIntegerInfix(node.Operator, l, right)
	case *Float:
		return evalFloatInfix(node.Operator, l, right)
	case *String:
		return evalStringInfix(node.Operator, l, right)
	case *Char:
		return evalStringInfix(node.Operator, &String{Value: l.Value}, right)
	case *Array:
		return evalArrayInfix(node.Operator, l, right)
	default:
		return nil, nil, &RuntimeError{Message: "unsupported infix operator: " + node.Operator}
	}
}

func evalIntegerInfix(op string, left *Integer, right Value) (Value, *Signal, error) {
	switch r := right.(type) {
	case *Integer:
		return evalNumericInfix(op, float64(left.Value), float64(r.Value), true)
	case *Float:
		return evalNumericInfix(op, float64(left.Value), r.Value, false)
	default:
		return nil, nil, &RuntimeError{Message: "type mismatch in integer operation"}
	}
}

func evalFloatInfix(op string, left *Float, right Value) (Value, *Signal, error) {
	switch r := right.(type) {
	case *Integer:
		return evalNumericInfix(op, left.Value, float64(r.Value), false)
	case *Float:
		return evalNumericInfix(op, left.Value, r.Value, false)
	default:
		return nil, nil, &RuntimeError{Message: "type mismatch in float operation"}
	}
}

func evalNumericInfix(op string, left, right float64, intResult bool) (Value, *Signal, error) {
	switch op {
	case "+":
		return numberValue(left+right, intResult), nil, nil
	case "-":
		return numberValue(left-right, intResult), nil, nil
	case "*":
		return numberValue(left*right, intResult), nil, nil
	case "/":
		return &Float{Value: left / right}, nil, nil
	case "%":
		if intResult {
			return &Integer{Value: int64(left) % int64(right)}, nil, nil
		}
		return nil, nil, &RuntimeError{Message: "modulo requires integers"}
	case "<":
		return &Boolean{Value: left < right}, nil, nil
	case "<=":
		return &Boolean{Value: left <= right}, nil, nil
	case ">":
		return &Boolean{Value: left > right}, nil, nil
	case ">=":
		return &Boolean{Value: left >= right}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unsupported numeric operator: " + op}
	}
}

func numberValue(val float64, intResult bool) Value {
	if intResult {
		return &Integer{Value: int64(val)}
	}
	return &Float{Value: val}
}

func evalStringInfix(op string, left *String, right Value) (Value, *Signal, error) {
	var r *String
	switch v := right.(type) {
	case *String:
		r = v
	case *Char:
		r = &String{Value: v.Value}
	default:
		return nil, nil, &RuntimeError{Message: "string operations require strings"}
	}
	switch op {
	case "+":
		return &String{Value: left.Value + r.Value}, nil, nil
	case "==":
		return &Boolean{Value: left.Value == r.Value}, nil, nil
	case "!=":
		return &Boolean{Value: left.Value != r.Value}, nil, nil
	case "<":
		return &Boolean{Value: left.Value < r.Value}, nil, nil
	case "<=":
		return &Boolean{Value: left.Value <= r.Value}, nil, nil
	case ">":
		return &Boolean{Value: left.Value > r.Value}, nil, nil
	case ">=":
		return &Boolean{Value: left.Value >= r.Value}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unsupported string operator: " + op}
	}
}

func evalArrayInfix(op string, left *Array, right Value) (Value, *Signal, error) {
	r, ok := right.(*Array)
	if !ok {
		return nil, nil, &RuntimeError{Message: "array operation requires array"}
	}
	switch op {
	case "+":
		out := append([]Value{}, left.Elements...)
		out = append(out, r.Elements...)
		return &Array{Elements: out}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unsupported array operator: " + op}
	}
}

func (e *Evaluator) evalAssignExpression(node *ast.AssignExpression, env *Environment) (Value, *Signal, error) {
	target, setter, err := e.resolveAssignable(node.Left, env)
	if err != nil {
		return nil, nil, err
	}
	right, sig, err := e.Eval(node.Right, env)
	if err != nil || sig != nil {
		return right, sig, err
	}

	var newVal Value
	switch node.Operator {
	case "=":
		newVal = right
	case "+=":
		newVal, err = e.applyBinary("+", target, right)
	case "-=":
		newVal, err = e.applyBinary("-", target, right)
	case "*=":
		newVal, err = e.applyBinary("*", target, right)
	case "/=":
		newVal, err = e.applyBinary("/", target, right)
	case "%=":
		newVal, err = e.applyBinary("%", target, right)
	default:
		return nil, nil, &RuntimeError{Message: "unknown assignment operator: " + node.Operator}
	}
	if err != nil {
		return nil, nil, err
	}
	setter(newVal)
	return newVal, nil, nil
}

func (e *Evaluator) evalPostfixExpression(node *ast.PostfixExpression, env *Environment) (Value, *Signal, error) {
	switch node.Operator {
	case "++", "--":
		target, setter, err := e.resolveAssignable(node.Left, env)
		if err != nil {
			return nil, nil, err
		}
		var delta float64 = 1
		if node.Operator == "--" {
			delta = -1
		}
		switch v := target.(type) {
		case *Integer:
			newVal := &Integer{Value: v.Value + int64(delta)}
			setter(newVal)
			return newVal, nil, nil
		case *Float:
			newVal := &Float{Value: v.Value + delta}
			setter(newVal)
			return newVal, nil, nil
		default:
			return nil, nil, &RuntimeError{Message: "increment/decrement requires number"}
		}
	default:
		return nil, nil, &RuntimeError{Message: "unknown postfix operator: " + node.Operator}
	}
}

func (e *Evaluator) evalAwaitExpression(node *ast.AwaitExpression, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.Eval(node.Value, env)
	if err != nil || sig != nil {
		return val, sig, err
	}
	task, ok := val.(*Task)
	if !ok {
		return nil, nil, &RuntimeError{Message: "wait expects task"}
	}
	return taskAwait(task)
}

func (e *Evaluator) evalRecoverExpression(node *ast.RecoverExpression, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.Eval(node.Target, env)
	if err == nil && sig == nil {
		return val, nil, nil
	}
	if sig != nil {
		return val, sig, err
	}
	re, ok := err.(*RecoverableError)
	if !ok {
		return nil, nil, err
	}
	fallbackEnv := NewEnclosedEnvironment(env)
	fallbackEnv.Define("error", recoverableErrorValue(re))
	return e.Eval(node.Fallback, fallbackEnv)
}

func (e *Evaluator) evalIfExpression(node *ast.IfExpression, env *Environment) (Value, *Signal, error) {
	cond, sig, err := e.Eval(node.Condition, env)
	if err != nil || sig != nil {
		return cond, sig, err
	}
	cb, ok := cond.(*Boolean)
	if !ok {
		return nil, nil, &RuntimeError{Message: "if condition must be bool"}
	}
	if cb.Value {
		return e.Eval(node.Consequence, env)
	}
	if node.Alternative != nil {
		return e.Eval(node.Alternative, env)
	}
	return UnitValue, nil, nil
}

func (e *Evaluator) evalBlockExpression(block *ast.BlockExpression, env *Environment) (Value, *Signal, error) {
	blockEnv := NewEnclosedEnvironment(env)
	var result Value = UnitValue
	for _, stmt := range block.Statements {
		val, sig, err := e.Eval(stmt, blockEnv)
		if err != nil || sig != nil {
			return val, sig, err
		}
		result = val
	}
	return result, nil, nil
}

func (e *Evaluator) evalMatchExpression(node *ast.MatchExpression, env *Environment) (Value, *Signal, error) {
	value, sig, err := e.Eval(node.Value, env)
	if err != nil || sig != nil {
		return value, sig, err
	}
	for _, arm := range node.Arms {
		armEnv := NewEnclosedEnvironment(env)
		ok, err := matchPattern(arm.Pattern, value, armEnv)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			continue
		}
		if arm.Guard != nil {
			guardVal, sig, err := e.Eval(arm.Guard, armEnv)
			if err != nil || sig != nil {
				return guardVal, sig, err
			}
			gb, ok := guardVal.(*Boolean)
			if !ok {
				return nil, nil, &RuntimeError{Message: "match guard must be bool"}
			}
			if !gb.Value {
				continue
			}
		}
		return e.Eval(arm.Body, armEnv)
	}
	return nil, nil, &RuntimeError{Message: "non-exhaustive match"}
}

func (e *Evaluator) evalForExpression(node *ast.ForExpression, env *Environment) (Value, *Signal, error) {
	loopEnv := NewEnclosedEnvironment(env)
	for _, binding := range node.Bindings {
		val, sig, err := e.Eval(binding.Value, loopEnv)
		if err != nil || sig != nil {
			return val, sig, err
		}
		ok, err := bindPattern(binding.Pattern, val, loopEnv)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, &RuntimeError{Message: "for binding pattern did not match"}
		}
	}

	for {
		condVal, sig, err := e.Eval(node.Condition, loopEnv)
		if err != nil || sig != nil {
			return condVal, sig, err
		}
		cb, ok := condVal.(*Boolean)
		if !ok {
			return nil, nil, &RuntimeError{Message: "for condition must be bool"}
		}
		if !cb.Value {
			break
		}

		_, sig, err = e.Eval(node.Body, loopEnv)
		if err != nil {
			return nil, nil, err
		}
		if sig != nil {
			switch sig.Type {
			case SignalContinue:
				continue
			case SignalBreak:
				if sig.Value != nil {
					return sig.Value, nil, nil
				}
				// break with no value: evaluate then block
				return e.evalThen(node, loopEnv)
			default:
				return nil, sig, nil
			}
		}
	}

	return e.evalThen(node, loopEnv)
}

func (e *Evaluator) evalThen(node *ast.ForExpression, env *Environment) (Value, *Signal, error) {
	if node.Then == nil {
		return UnitValue, nil, nil
	}
	return e.Eval(node.Then, env)
}

func (e *Evaluator) evalCallExpression(node *ast.CallExpression, env *Environment) (Value, *Signal, error) {
	function, sig, err := e.Eval(node.Function, env)
	if err != nil || sig != nil {
		return function, sig, err
	}

	args := make([]Value, 0, len(node.Arguments))
	hasPlaceholder := false
	for _, arg := range node.Arguments {
		if _, ok := arg.(*ast.Placeholder); ok {
			args = append(args, nil)
			hasPlaceholder = true
			continue
		}
		val, sig, err := e.Eval(arg, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		args = append(args, val)
	}

	if hasPlaceholder {
		return &Partial{Target: function, Args: args}, nil, nil
	}
	return e.applyFunction(function, args)
}

func (e *Evaluator) evalMemberExpression(node *ast.MemberExpression, env *Environment) (Value, *Signal, error) {
	object, sig, err := e.Eval(node.Object, env)
	if err != nil || sig != nil {
		return object, sig, err
	}

	switch obj := object.(type) {
	case *Object:
		val, ok := obj.Pairs[node.Property.Value]
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + node.Property.Value}
		}
		return val, nil, nil
	case *ModuleObject:
		if obj.Env == nil {
			return nil, nil, &RuntimeError{Message: "member access on invalid module object"}
		}
		val, ok := obj.Env.GetLocal(node.Property.Value)
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + node.Property.Value}
		}
		return val, nil, nil
	case *Array:
		if node.Property.Value == "length" {
			return &Integer{Value: int64(len(obj.Elements))}, nil, nil
		}
		return e.arrayMethod(obj, node.Property.Value)
	case *String:
		if node.Property.Value == "length" {
			return &Integer{Value: int64(utf8.RuneCountInString(obj.Value))}, nil, nil
		}
		return e.stringMethod(obj, node.Property.Value)
	case *Map:
		return e.mapMethod(obj, node.Property.Value)
	case *Set:
		if node.Property.Value == "size" {
			return &Integer{Value: int64(len(obj.Elements))}, nil, nil
		}
		return e.setMethod(obj, node.Property.Value)
	case *Channel:
		return e.channelMethod(obj, node.Property.Value)
	case *Task:
		return e.taskMethod(obj, node.Property.Value)
	default:
		if object == nil {
			return nil, nil, &RuntimeError{Message: "member access on non-object (got <nil>)"}
		}
		return nil, nil, &RuntimeError{Message: fmt.Sprintf("member access on non-object (%s.%s)", object.Type(), node.Property.Value)}
	}
}

func (e *Evaluator) evalIndexExpression(node *ast.IndexExpression, env *Environment) (Value, *Signal, error) {
	left, sig, err := e.Eval(node.Left, env)
	if err != nil || sig != nil {
		return left, sig, err
	}
	indexVal, sig, err := e.Eval(node.Index, env)
	if err != nil || sig != nil {
		return indexVal, sig, err
	}

	arr, ok := left.(*Array)
	if !ok {
		return nil, nil, &RuntimeError{Message: "indexing requires array"}
	}
	idx, ok := indexVal.(*Integer)
	if !ok {
		return nil, nil, &RuntimeError{Message: "index must be integer"}
	}
	i := int(idx.Value)
	if i < 0 || i >= len(arr.Elements) {
		return nil, nil, &RuntimeError{Message: "index out of bounds"}
	}
	return arr.Elements[i], nil, nil
}

func (e *Evaluator) evalSliceExpression(node *ast.SliceExpression, env *Environment) (Value, *Signal, error) {
	left, sig, err := e.Eval(node.Left, env)
	if err != nil || sig != nil {
		return left, sig, err
	}
	arr, ok := left.(*Array)
	if !ok {
		return nil, nil, &RuntimeError{Message: "slice requires array"}
	}

	start := 0
	end := len(arr.Elements)

	if node.Start != nil {
		val, sig, err := e.Eval(node.Start, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		idx, ok := val.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "slice start must be integer"}
		}
		start = normalizeIndex(int(idx.Value), len(arr.Elements))
	}

	if node.End != nil {
		val, sig, err := e.Eval(node.End, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		idx, ok := val.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "slice end must be integer"}
		}
		end = normalizeIndex(int(idx.Value), len(arr.Elements))
	}

	if start < 0 || start > len(arr.Elements) || end < 0 || end > len(arr.Elements) {
		return nil, nil, &RuntimeError{Message: "slice bounds out of range"}
	}
	if start >= end {
		return &Array{Elements: []Value{}}, nil, nil
	}
	out := append([]Value{}, arr.Elements[start:end]...)
	return &Array{Elements: out}, nil, nil
}

func normalizeIndex(idx int, length int) int {
	if idx < 0 {
		return length + idx
	}
	return idx
}

func (e *Evaluator) evalArrayLiteral(node *ast.ArrayLiteral, env *Environment) (Value, *Signal, error) {
	elements := make([]Value, 0, len(node.Elements))
	for _, el := range node.Elements {
		val, sig, err := e.Eval(el, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		elements = append(elements, val)
	}
	return &Array{Elements: elements}, nil, nil
}

func (e *Evaluator) evalObjectLiteral(node *ast.ObjectLiteral, env *Environment) (Value, *Signal, error) {
	obj := &Object{Pairs: make(map[string]Value)}
	for _, entry := range node.Entries {
		if entry.Spread {
			val, sig, err := e.Eval(entry.Value, env)
			if err != nil || sig != nil {
				return val, sig, err
			}
			pairs, ok := objectPairs(val)
			if !ok {
				return nil, nil, &RuntimeError{Message: "object spread requires object"}
			}
			for k, v := range pairs {
				obj.Pairs[k] = v
			}
			continue
		}

		val, sig, err := e.Eval(entry.Value, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		obj.Pairs[entry.Key] = val
	}
	return obj, nil, nil
}

func (e *Evaluator) evalStructInitExpression(node *ast.StructInitExpression, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.evalObjectLiteral(node.Value, env)
	if err != nil || sig != nil {
		return val, sig, err
	}
	return val, nil, nil
}

func (e *Evaluator) evalRangeExpression(node *ast.RangeExpression, env *Environment) (Value, *Signal, error) {
	startVal, sig, err := e.Eval(node.Start, env)
	if err != nil || sig != nil {
		return startVal, sig, err
	}
	endVal, sig, err := e.Eval(node.End, env)
	if err != nil || sig != nil {
		return endVal, sig, err
	}
	if _, ok := startVal.(*Float); ok {
		return nil, nil, &RuntimeError{Message: "float ranges are not allowed"}
	}
	if _, ok := endVal.(*Float); ok {
		return nil, nil, &RuntimeError{Message: "float ranges are not allowed"}
	}

	stepVal := Value(&Integer{Value: 1})
	if node.Step != nil {
		stepVal, sig, err = e.Eval(node.Step, env)
		if err != nil || sig != nil {
			return stepVal, sig, err
		}
	}

	return buildRange(startVal, endVal, stepVal)
}

func buildRange(start, end, step Value) (Value, *Signal, error) {
	switch s := start.(type) {
	case *Integer:
		e, ok := end.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "range endpoints must match types"}
		}
		st, ok := step.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "integer range step must be integer"}
		}
		if st.Value == 0 {
			return nil, nil, &RuntimeError{Message: "range step cannot be zero"}
		}
		return buildIntRange(s.Value, e.Value, st.Value), nil, nil
	case *Float:
		e, ok := end.(*Float)
		if !ok {
			return nil, nil, &RuntimeError{Message: "range endpoints must match types"}
		}
		var stepVal float64
		switch st := step.(type) {
		case *Float:
			stepVal = st.Value
		case *Integer:
			stepVal = float64(st.Value)
		default:
			return nil, nil, &RuntimeError{Message: "float range step must be number"}
		}
		if stepVal == 0 {
			return nil, nil, &RuntimeError{Message: "range step cannot be zero"}
		}
		return buildFloatRange(s.Value, e.Value, stepVal), nil, nil
	case *Char:
		e, ok := end.(*Char)
		if !ok {
			return nil, nil, &RuntimeError{Message: "range endpoints must match types"}
		}
		st, ok := step.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "char range step must be integer"}
		}
		if st.Value == 0 {
			return nil, nil, &RuntimeError{Message: "range step cannot be zero"}
		}
		return buildCharRange(s.Value, e.Value, st.Value), nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unsupported range endpoint type"}
	}
}

func buildIntRange(start, end, step int64) Value {
	out := []Value{}
	if step > 0 {
		for i := start; i <= end; i += step {
			out = append(out, &Integer{Value: i})
		}
	} else {
		for i := start; i >= end; i += step {
			out = append(out, &Integer{Value: i})
		}
	}
	return &Array{Elements: out}
}

func buildFloatRange(start, end, step float64) Value {
	out := []Value{}
	if step > 0 {
		for i := start; i < end; i += step {
			out = append(out, &Float{Value: i})
		}
	} else {
		for i := start; i > end; i += step {
			out = append(out, &Float{Value: i})
		}
	}
	return &Array{Elements: out}
}

func buildCharRange(start, end string, step int64) Value {
	sr, _ := utf8.DecodeRuneInString(start)
	er, _ := utf8.DecodeRuneInString(end)
	out := []Value{}
	if step > 0 {
		for r := sr; r <= er; r += rune(step) {
			out = append(out, &Char{Value: string(r)})
		}
	} else {
		for r := sr; r >= er; r += rune(step) {
			out = append(out, &Char{Value: string(r)})
		}
	}
	return &Array{Elements: out}
}

func (e *Evaluator) evalQueryExpression(node *ast.QueryExpression, env *Environment) (Value, *Signal, error) {
	sourceVal, sig, err := e.Eval(node.Source, env)
	if err != nil || sig != nil {
		return sourceVal, sig, err
	}
	source, ok := sourceVal.(*Array)
	if !ok {
		return nil, nil, &RuntimeError{Message: "query source must be array"}
	}

	rows := []queryRow{}

	for _, item := range source.Elements {
		rowEnv := NewEnclosedEnvironment(env)
		rowEnv.Define(node.Var.Value, item)

		keep := true
		for _, whereExpr := range node.Where {
			val, sig, err := e.Eval(whereExpr, rowEnv)
			if err != nil || sig != nil {
				return val, sig, err
			}
			b, ok := val.(*Boolean)
			if !ok {
				return nil, nil, &RuntimeError{Message: "query where must be bool"}
			}
			if !b.Value {
				keep = false
				break
			}
		}

		if !keep {
			continue
		}

		var key Value
		if node.OrderBy != nil {
			val, sig, err := e.Eval(node.OrderBy, rowEnv)
			if err != nil || sig != nil {
				return val, sig, err
			}
			key = val
		}
		rows = append(rows, queryRow{item: item, key: key})
	}

	if node.OrderBy != nil {
		if err := sortRows(rows); err != nil {
			return nil, nil, err
		}
	}

	results := []Value{}
	for _, r := range rows {
		rowEnv := NewEnclosedEnvironment(env)
		rowEnv.Define(node.Var.Value, r.item)
		val, sig, err := e.Eval(node.Select, rowEnv)
		if err != nil || sig != nil {
			return val, sig, err
		}
		results = append(results, val)
	}
	return &Array{Elements: results}, nil, nil
}

func sortRows(rows []queryRow) error {
	if len(rows) == 0 {
		return nil
	}
	for _, row := range rows {
		if row.key == nil {
			return &RuntimeError{Message: "orderby requires comparable key"}
		}
	}

	switch rows[0].key.(type) {
	case *Integer, *Float, *String:
	default:
		return &RuntimeError{Message: "orderby key must be int, float, or string"}
	}
	firstType := rows[0].key.Type()
	for _, row := range rows {
		if row.key.Type() != firstType {
			return &RuntimeError{Message: "orderby keys must be the same type"}
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		return compareForSort(rows[i].key, rows[j].key)
	})
	return nil
}

func compareForSort(left, right Value) bool {
	switch l := left.(type) {
	case *Integer:
		return l.Value < right.(*Integer).Value
	case *Float:
		return l.Value < right.(*Float).Value
	case *String:
		return l.Value < right.(*String).Value
	default:
		return false
	}
}

func (e *Evaluator) evalRaceExpression(node *ast.RaceExpression, env *Environment) (Value, *Signal, error) {
	tasks := []*Task{}
	for _, taskExpr := range node.Tasks {
		task, err := e.spawnTask(taskExpr, env)
		if err != nil {
			return nil, nil, err
		}
		tasks = append(tasks, task)
	}
	raceTask := newTask()
	go func() {
		results := make(chan taskResult, len(tasks))
		for _, task := range tasks {
			go func(t *Task) {
				val, sig, err := taskAwait(t)
				if err != nil {
					results <- taskResult{value: nil, err: err}
					return
				}
				if sig != nil {
					results <- taskResult{value: nil, err: &RuntimeError{Message: "break/continue outside loop"}}
					return
				}
				results <- taskResult{value: val, err: err}
			}(task)
		}
		first := <-results
		raceTask.complete(first.value, first.err)
	}()
	return raceTask, nil, nil
}

func (e *Evaluator) evalSpawnExpression(node *ast.SpawnExpression, env *Environment) (Value, *Signal, error) {
	if node.Task != nil {
		task, err := e.spawnTask(node.Task, env)
		if err != nil {
			return nil, nil, err
		}
		return task, nil, nil
	}

	tasks := []*Task{}
	for _, expr := range node.Group {
		task, err := e.spawnTask(expr, env)
		if err != nil {
			return nil, nil, err
		}
		tasks = append(tasks, task)
	}
	join := newTask()
	go func() {
		results := make([]Value, len(tasks))
		for i, t := range tasks {
			val, sig, err := taskAwait(t)
			if err != nil {
				join.complete(nil, err)
				return
			}
			if sig != nil {
				join.complete(nil, &RuntimeError{Message: "break/continue outside loop"})
				return
			}
			results[i] = val
		}
		join.complete(&Array{Elements: results}, nil)
	}()
	return join, nil, nil
}

func (e *Evaluator) evalBreakExpression(node *ast.BreakExpression, env *Environment) (Value, *Signal, error) {
	if node.Value == nil {
		return UnitValue, &Signal{Type: SignalBreak}, nil
	}
	val, sig, err := e.Eval(node.Value, env)
	if err != nil || sig != nil {
		return val, sig, err
	}
	return val, &Signal{Type: SignalBreak, Value: val}, nil
}

func (e *Evaluator) spawnTask(expr ast.Expression, env *Environment) (*Task, error) {
	task := newTask()
	go func() {
		val, sig, err := e.Eval(expr, env)
		if err != nil {
			if exitErr, ok := err.(*ExitError); ok {
				exitProcess(exitErr.Message)
				return
			}
			exitProcess(e.formatError(err))
			return
		}
		if sig != nil {
			task.complete(nil, &RuntimeError{Message: "break/continue outside loop"})
			return
		}
		task.complete(val, nil)
	}()
	return task, nil
}

func (e *Evaluator) resolveAssignable(node ast.Expression, env *Environment) (Value, func(Value), error) {
	switch n := node.(type) {
	case *ast.Identifier:
		val, ok := env.Get(n.Value)
		if !ok {
			return nil, nil, &RuntimeError{Message: "undefined identifier: " + n.Value}
		}
		return val, func(v Value) { env.Set(n.Value, v) }, nil
	case *ast.MemberExpression:
		objVal, sig, err := e.Eval(n.Object, env)
		if err != nil || sig != nil {
			return nil, nil, err
		}
		switch obj := objVal.(type) {
		case *Object:
			return obj.Pairs[n.Property.Value], func(v Value) { obj.Pairs[n.Property.Value] = v }, nil
		case *ModuleObject:
			if obj.Env == nil {
				return nil, nil, &RuntimeError{Message: "member assignment requires object"}
			}
			val, _ := obj.Env.GetLocal(n.Property.Value)
			return val, func(v Value) { obj.Env.Define(n.Property.Value, v) }, nil
		default:
			return nil, nil, &RuntimeError{Message: "member assignment requires object"}
		}
	case *ast.IndexExpression:
		left, sig, err := e.Eval(n.Left, env)
		if err != nil || sig != nil {
			return nil, nil, err
		}
		arr, ok := left.(*Array)
		if !ok {
			return nil, nil, &RuntimeError{Message: "index assignment requires array"}
		}
		indexVal, sig, err := e.Eval(n.Index, env)
		if err != nil || sig != nil {
			return nil, nil, err
		}
		idx, ok := indexVal.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "index must be integer"}
		}
		i := int(idx.Value)
		if i < 0 || i >= len(arr.Elements) {
			return nil, nil, &RuntimeError{Message: "index out of bounds"}
		}
		return arr.Elements[i], func(v Value) { arr.Elements[i] = v }, nil
	default:
		return nil, nil, &RuntimeError{Message: "invalid assignment target"}
	}
}

func (e *Evaluator) applyBinary(op string, left, right Value) (Value, error) {
	switch op {
	case "+":
		switch l := left.(type) {
		case *Integer:
			val, _, err := evalIntegerInfix(op, l, right)
			return val, err
		case *Float:
			val, _, err := evalFloatInfix(op, l, right)
			return val, err
		case *String:
			val, _, err := evalStringInfix(op, l, right)
			return val, err
		case *Array:
			val, _, err := evalArrayInfix(op, l, right)
			return val, err
		default:
			return nil, &RuntimeError{Message: "unsupported +="}
		}
	case "-", "*", "/", "%":
		switch l := left.(type) {
		case *Integer:
			val, _, err := evalIntegerInfix(op, l, right)
			return val, err
		case *Float:
			val, _, err := evalFloatInfix(op, l, right)
			return val, err
		default:
			return nil, &RuntimeError{Message: "unsupported assignment operator"}
		}
	default:
		return nil, &RuntimeError{Message: "unknown assignment operator"}
	}
}

func (e *Evaluator) applyFunction(fn Value, args []Value) (Value, *Signal, error) {
	switch f := fn.(type) {
	case *Builtin:
		val, err := f.Fn(e, args)
		return val, nil, err
	case *Function:
		if len(args) != len(f.Params) {
			return nil, nil, &RuntimeError{Message: "wrong number of arguments"}
		}
		extended := NewEnclosedEnvironment(f.Env)
		for i, param := range f.Params {
			ok, err := bindPattern(param, args[i], extended)
			if err != nil {
				return nil, nil, err
			}
			if !ok {
				return nil, nil, &RuntimeError{Message: "parameter pattern did not match"}
			}
		}
		val, sig, err := e.Eval(f.Body, extended)
		if err != nil {
			return nil, nil, err
		}
		if sig != nil {
			return nil, nil, &RuntimeError{Message: "break/continue outside loop"}
		}
		return val, nil, nil
	case *Partial:
		filled := []Value{}
		argIndex := 0
		for _, arg := range f.Args {
			if arg == nil {
				if argIndex >= len(args) {
					return nil, nil, &RuntimeError{Message: "not enough arguments for partial"}
				}
				filled = append(filled, args[argIndex])
				argIndex++
				continue
			}
			filled = append(filled, arg)
		}
		if argIndex != len(args) {
			return nil, nil, &RuntimeError{Message: "too many arguments for partial"}
		}
		return e.applyFunction(f.Target, filled)
	default:
		return nil, nil, &RuntimeError{Message: "not a function"}
	}
}

func (e *Evaluator) arrayMethod(arr *Array, name string) (Value, *Signal, error) {
	switch name {
	case "map", "filter", "reduce", "sum", "find", "sort":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, arr)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown array member: " + name}
	}
}

func (e *Evaluator) channelMethod(ch *Channel, name string) (Value, *Signal, error) {
	switch name {
	case "send", "recv", "done":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, ch)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown channel member: " + name}
	}
}

func (e *Evaluator) stringMethod(str *String, name string) (Value, *Signal, error) {
	switch name {
	case "split", "chars", "trim", "toLower", "toUpper", "contains", "startsWith", "endsWith", "replace":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, str)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown string member: " + name}
	}
}

func (e *Evaluator) mapMethod(m *Map, name string) (Value, *Signal, error) {
	switch name {
	case "get", "set", "has", "delete", "keys", "values":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, m)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown map member: " + name}
	}
}

func (e *Evaluator) setMethod(s *Set, name string) (Value, *Signal, error) {
	switch name {
	case "add", "has", "delete", "values":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, s)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown set member: " + name}
	}
}

func (e *Evaluator) taskMethod(t *Task, name string) (Value, *Signal, error) {
	switch name {
	case "then":
		builtin := getBuiltin(name)
		if builtin == nil {
			return nil, nil, &RuntimeError{Message: "unknown builtin: " + name}
		}
		return &Builtin{Name: name, Fn: bindReceiver(builtin.Fn, t)}, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unknown task member: " + name}
	}
}

func bindPattern(pattern ast.Pattern, value Value, env *Environment) (bool, error) {
	return matchPattern(pattern, value, env)
}
