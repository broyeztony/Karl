package interpreter

import (
	"fmt"
	"karl/ast"
)

type Evaluator struct {
	source      string
	filename    string
	projectRoot string
	modules     *moduleState

	runtime     *runtimeState
	currentTask *Task
}

func NewEvaluator() *Evaluator {
	return &Evaluator{modules: newModuleState(), runtime: newRuntimeState()}
}

func NewEvaluatorWithSource(source string) *Evaluator {
	return &Evaluator{source: source, modules: newModuleState(), runtime: newRuntimeState()}
}

func NewEvaluatorWithSourceAndFilename(source string, filename string) *Evaluator {
	return &Evaluator{source: source, filename: filename, modules: newModuleState(), runtime: newRuntimeState()}
}

func NewEvaluatorWithSourceFilenameAndRoot(source string, filename string, root string) *Evaluator {
	return &Evaluator{
		source:      source,
		filename:    filename,
		projectRoot: root,
		modules:     newModuleState(),
		runtime:     newRuntimeState(),
	}
}

func (e *Evaluator) SetProjectRoot(root string) {
	e.projectRoot = root
}

func (e *Evaluator) SetTaskFailurePolicy(policy string) error {
	if e.runtime == nil {
		e.runtime = newRuntimeState()
	}
	return e.runtime.setTaskFailurePolicy(policy)
}

func (e *Evaluator) cloneForTask(task *Task) *Evaluator {
	return &Evaluator{
		source:      e.source,
		filename:    e.filename,
		projectRoot: e.projectRoot,
		modules:     e.modules,
		runtime:     e.runtime,
		currentTask: task,
	}
}

func (e *Evaluator) newTask(parent *Task, internal bool) *Task {
	if e.runtime == nil {
		e.runtime = newRuntimeState()
	}
	t := newTask()
	t.internal = internal
	t.parent = parent
	t.source = e.source
	t.filename = e.filename
	if parent != nil {
		parent.addChild(t)
	}
	e.runtime.registerTask(t)
	return t
}

func errorValue(err error) Value {
	switch e := err.(type) {
	case *RecoverableError:
		kind := e.Kind
		if kind == "" {
			kind = "error"
		}
		return &Object{Pairs: map[string]Value{
			"kind":    &String{Value: kind},
			"message": &String{Value: e.Message},
		}}
	case *RuntimeError:
		return &Object{Pairs: map[string]Value{
			"kind":    &String{Value: "runtime"},
			"message": &String{Value: e.Message},
		}}
	default:
		return &Object{Pairs: map[string]Value{
			"kind":    &String{Value: "error"},
			"message": &String{Value: err.Error()},
		}}
	}
}

func (e *Evaluator) Eval(node ast.Node, env *Environment) (Value, *Signal, error) {
	if e.runtime != nil {
		if err := e.runtime.getFatalTaskFailure(); err != nil {
			return nil, nil, err
		}
	}
	if e.currentTask != nil && e.currentTask.canceled() {
		return nil, nil, canceledError()
	}

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
	if err == nil && sig == nil && e.runtime != nil {
		if fatalErr := e.runtime.getFatalTaskFailure(); fatalErr != nil {
			return nil, nil, fatalErr
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
		// Support truthy/falsy evaluation for negation
		return &Boolean{Value: !isTruthy(right)}, nil, nil
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
		// Support truthy/falsy evaluation for logical operators
		leftTruthy := isTruthy(left)
		if node.Operator == "&&" && !leftTruthy {
			return &Boolean{Value: false}, nil, nil
		}
		if node.Operator == "||" && leftTruthy {
			return &Boolean{Value: true}, nil, nil
		}
		right, sig, err := e.Eval(node.Right, env)
		if err != nil || sig != nil {
			return right, sig, err
		}
		return &Boolean{Value: isTruthy(right)}, nil, nil
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
