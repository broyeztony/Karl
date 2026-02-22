package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

type Value = interpreter.Value
type Integer = interpreter.Integer
type Boolean = interpreter.Boolean
type String = interpreter.String
type Float = interpreter.Float
type Array = interpreter.Array
type Object = interpreter.Object

var NullValue = interpreter.NullValue
var Equivalent = interpreter.Equivalent

func evalInput(t *testing.T, input string) (Value, error) {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.Fatalf("parse failed")
	}

	eval := interpreter.NewEvaluator()
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if sig != nil {
		return nil, fmt.Errorf("unexpected signal: %v", sig.Type)
	}
	return val, err
}

func mustEval(t *testing.T, input string) Value {
	t.Helper()
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	return val
}

func assertInteger(t *testing.T, val Value, expected int64) {
	t.Helper()
	i, ok := val.(*Integer)
	if !ok {
		t.Fatalf("expected Integer, got %T (%v)", val, val)
	}
	if i.Value != expected {
		t.Fatalf("expected %d, got %d", expected, i.Value)
	}
}

func assertString(t *testing.T, val Value, expected string) {
	t.Helper()
	s, ok := val.(*String)
	if !ok {
		t.Fatalf("expected String, got %T (%v)", val, val)
	}
	if s.Value != expected {
		t.Fatalf("expected %q, got %q", expected, s.Value)
	}
}

func assertFloat(t *testing.T, val Value, expected float64) {
	t.Helper()
	f, ok := val.(*Float)
	if !ok {
		t.Fatalf("expected Float, got %T (%v)", val, val)
	}
	if f.Value != expected {
		t.Fatalf("expected %g, got %g", expected, f.Value)
	}
}

func assertEquivalent(t *testing.T, val Value, expected Value) {
	t.Helper()
	if !Equivalent(val, expected) {
		t.Fatalf("expected %s, got %s", expected.Inspect(), val.Inspect())
	}
}

func TestEvalArithmetic(t *testing.T) {
	val := mustEval(t, "1 + 2 * 3")
	assertInteger(t, val, 7)
}

func TestEvalLetAndIdentifier(t *testing.T) {
	val := mustEval(t, "let x = 5; x")
	assertInteger(t, val, 5)
}

func TestEvalIf(t *testing.T) {
	val := mustEval(t, "if true { 1 } else { 2 }")
	assertInteger(t, val, 1)
}

func TestEvalMatch(t *testing.T) {
	val := mustEval(t, "match 2 { case 1 -> 10 case 2 -> 20 }")
	assertInteger(t, val, 20)
}

func TestEvalForBreakValue(t *testing.T) {
	input := `
let found = for true with i = 0, acc = 0 {
    i++
    acc += i
    if i == 3 { break acc }
} then 0
found
`
	val := mustEval(t, input)
	assertInteger(t, val, 6)
}

func TestEvalForBreakNoValue(t *testing.T) {
	input := `
let result = for true with i = 0 {
    i++
    if i == 2 { break }
} then i
result
`
	val := mustEval(t, input)
	assertInteger(t, val, 2)
}

func TestEvalRangeInt(t *testing.T) {
	val := mustEval(t, "1..3")
	expected := &Array{Elements: []Value{
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalRangeFloatRejected(t *testing.T) {
	p := parser.New(lexer.New("1.0..3.0"))
	_ = p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatalf("expected parse error")
	}
}

func TestEvalSliceNegative(t *testing.T) {
	val := mustEval(t, "let xs = [1, 2, 3, 4]; xs[..-1]")
	expected := &Array{Elements: []Value{
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalStringSlice(t *testing.T) {
	val := mustEval(t, `let s = "some string"; [s[0..4], s[5..], s[..4], s[..-1], s[-3..]]`)
	expected := &Array{Elements: []Value{
		&String{Value: "some"},
		&String{Value: "string"},
		&String{Value: "some"},
		&String{Value: "some strin"},
		&String{Value: "ing"},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalStringSliceUnicode(t *testing.T) {
	val := mustEval(t, `let s = "héll🙂"; s[1..4]`)
	assertString(t, val, "éll")
}

func TestEvalStringSliceEmptyWhenStartAtOrPastEnd(t *testing.T) {
	val := mustEval(t, `let s = "abc"; [s[2..2], s[2..1]]`)
	expected := &Array{Elements: []Value{
		&String{Value: ""},
		&String{Value: ""},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalStringSliceOutOfRange(t *testing.T) {
	_, err := evalInput(t, `let s = "abc"; s[0..5]`)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "slice bounds out of range") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvalEqv(t *testing.T) {
	val := mustEval(t, "let a = [1, 2]; let b = [1, 2]; [a == b, a eqv b]")
	expected := &Array{Elements: []Value{
		&Boolean{Value: false},
		&Boolean{Value: true},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalParseIntInvalid(t *testing.T) {
	_, err := evalInput(t, `parseInt("nope")`)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestEvalStrBuiltin(t *testing.T) {
	val := mustEval(t, `str(123)`)
	assertString(t, val, "123")
}

func TestEvalLenBuiltin(t *testing.T) {
	val := mustEval(t, `
let m = map()
m.set("a", 1)
let s = set([1, 2, 2])
let o = { x: 1, y: 2, }
let arr = [1, 2]
let a = len("hé")
let b = len(arr)
let c = len(m)
let d = len(s)
let e = len(o)
let out = [a, b, c, d, e]
out
`)
	expected := &Array{Elements: []Value{
		&Integer{Value: 2},
		&Integer{Value: 2},
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 2},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalLenBuiltinRejectsUnsupportedType(t *testing.T) {
	_, err := evalInput(t, `len(1)`)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "len expects string, array, map, set, or object") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvalEncodeDecodeJSON(t *testing.T) {
	val := mustEval(t, `jsonDecode(jsonEncode({ a: 1, b: [true, null, "x"] }))`)
	expected := &Object{Pairs: map[string]Value{
		"a": &Integer{Value: 1},
		"b": &Array{Elements: []Value{
			&Boolean{Value: true},
			NullValue,
			&String{Value: "x"},
		}},
	}}
	assertEquivalent(t, val, expected)
}

func TestTruthyFalsyIf(t *testing.T) {
	input := `
let m = map()
let s = set()
let m2 = map()
m2.set("a", 1)
let s2 = set()
s2.add(1)
let a = if null { 1 } else { 0 }
let b = if false { 1 } else { 0 }
let c = if 0 { 1 } else { 0 }
let d = if 0.0 { 1 } else { 0 }
let e = if "" { 1 } else { 0 }
let f = if [] { 1 } else { 0 }
let g = if {} { 1 } else { 0 }
let h = if m { 1 } else { 0 }
let i = if s { 1 } else { 0 }
let j = if 1 { 1 } else { 0 }
let k = if -1 { 1 } else { 0 }
let l = if 1.5 { 1 } else { 0 }
let m3 = if "x" { 1 } else { 0 }
let n = if [1] { 1 } else { 0 }
let o = if { a: 1 } { 1 } else { 0 }
let p = if m2 { 1 } else { 0 }
let q = if s2 { 1 } else { 0 };
[a, b, c, d, e, f, g, h, i, j, k, l, m3, n, o, p, q]
`
	val := mustEval(t, input)
	expected := &Array{Elements: []Value{
		&Integer{Value: 0},
		&Integer{Value: 0},
		&Integer{Value: 0},
		&Integer{Value: 0},
		&Integer{Value: 0},
		&Integer{Value: 0},
		&Integer{Value: 0},
		&Integer{Value: 0},
		&Integer{Value: 0},
		&Integer{Value: 1},
		&Integer{Value: 1},
		&Integer{Value: 1},
		&Integer{Value: 1},
		&Integer{Value: 1},
		&Integer{Value: 1},
		&Integer{Value: 1},
		&Integer{Value: 1},
	}}
	assertEquivalent(t, val, expected)
}

func TestTruthyFalsyNegation(t *testing.T) {
	val := mustEval(t, `[!null, !0, !"", !42]`)
	expected := &Array{Elements: []Value{
		&Boolean{Value: true},
		&Boolean{Value: true},
		&Boolean{Value: true},
		&Boolean{Value: false},
	}}
	assertEquivalent(t, val, expected)
}

func TestTruthyFalsyLogicalOperators(t *testing.T) {
	val := mustEval(t, `[0 || 1, 1 || 0, 0 && 1, 1 && 2]`)
	expected := &Array{Elements: []Value{
		&Boolean{Value: true},
		&Boolean{Value: true},
		&Boolean{Value: false},
		&Boolean{Value: true},
	}}
	assertEquivalent(t, val, expected)
}

func TestTruthyFalsyShortCircuit(t *testing.T) {
	val := mustEval(t, `let a = false && fail("boom"); let b = true || fail("boom"); [a, b]`)
	expected := &Array{Elements: []Value{
		&Boolean{Value: false},
		&Boolean{Value: true},
	}}
	assertEquivalent(t, val, expected)
}

func TestTruthyFalsyForCondition(t *testing.T) {
	val := mustEval(t, `
let iterations = for count with count = 2, iter = 0 {
    iter = iter + 1
    count = count - 1
} then iter
iterations
`)
	assertInteger(t, val, 2)
}

func TestTruthyFalsyMatchGuard(t *testing.T) {
	val := mustEval(t, `[match 1 { case _ if 0 -> 1 case _ -> 2 }, match 1 { case _ if 1 -> 3 case _ -> 4 }]`)
	expected := &Array{Elements: []Value{
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalObjectStringIndexRead(t *testing.T) {
	val := mustEval(t, `let obj = jsonDecode("{\"a-field\": 42}"); obj["a-field"]`)
	assertInteger(t, val, 42)
}

func TestEvalObjectStringIndexAssignment(t *testing.T) {
	val := mustEval(t, `let obj = {}; obj["a-field"] = 3; obj["a-field"] += 4; obj["a-field"]`)
	assertInteger(t, val, 7)
}

func TestEvalObjectCharIndex(t *testing.T) {
	val := mustEval(t, `let obj = {}; obj['x'] = 1; obj["x"]`)
	assertInteger(t, val, 1)
}

func TestEvalObjectStringIndexMissingProperty(t *testing.T) {
	_, err := evalInput(t, `let obj = {}; obj["missing"]`)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "missing property: missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvalObjectStringIndexRequiresString(t *testing.T) {
	_, err := evalInput(t, `let obj = { a: 1 }; obj[0]`)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "object index must be string or char") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvalArrayIndexStillRequiresInteger(t *testing.T) {
	_, err := evalInput(t, `let arr = [1, 2, 3]; arr["0"]`)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "index must be integer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvalDecodeJSONOverflow(t *testing.T) {
	_, err := evalInput(t, `jsonDecode("999999999999999999999")`)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestEvalRecoverDecodeJSON(t *testing.T) {
	val := mustEval(t, `jsonDecode("{\"foo\": }") ? { foo: "bar", }`)
	expected := &Object{Pairs: map[string]Value{
		"foo": &String{Value: "bar"},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalFailRecover(t *testing.T) {
	val := mustEval(t, `fail("nope") ? { "fallback" }`)
	assertString(t, val, "fallback")
}

func TestEvalRecoverRuntimeErrorMemberAccess(t *testing.T) {
	val := mustEval(t, `let obj = {}; (obj.missing) ? { "fallback" }`)
	assertString(t, val, "fallback")
}

func TestEvalRecoverRuntimeErrorIndexAccess(t *testing.T) {
	val := mustEval(t, `let obj = {}; obj["missing"] ? { "fallback" }`)
	assertString(t, val, "fallback")
}

func TestEvalRecoverRuntimeErrorDirectFallback(t *testing.T) {
	val := mustEval(t, `let obj = {}; obj.missing ? "fallback"`)
	assertString(t, val, "fallback")
}

func TestEvalRecoverDirectExpressionFallback(t *testing.T) {
	val := mustEval(t, `fail("nope") ? 1 + 2`)
	assertInteger(t, val, 3)
}

func TestEvalSetBasics(t *testing.T) {
	val := mustEval(t, `let s = set(); s.add(1); s.add(2); s.add(2); [s.has(1), s.has(3), s.size]`)
	expected := &Array{Elements: []Value{
		&Boolean{Value: true},
		&Boolean{Value: false},
		&Integer{Value: 2},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalSort(t *testing.T) {
	val := mustEval(t, `sort([3, 1, 2], (a, b) -> a - b)`)
	expected := &Array{Elements: []Value{
		&Integer{Value: 1},
		&Integer{Value: 2},
		&Integer{Value: 3},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalRandIntDeterministic(t *testing.T) {
	val := mustEval(t, `randInt(5, 5)`)
	assertInteger(t, val, 5)
}

func TestEvalRandFloatDeterministic(t *testing.T) {
	val := mustEval(t, `randFloat(1.25, 1.25)`)
	assertFloat(t, val, 1.25)
}

func TestEvalFileIO(t *testing.T) {
	dir := t.TempDir()
	path := filepath.ToSlash(filepath.Join(dir, "data.txt"))
	input := fmt.Sprintf(`writeFile("%s", "hi"); appendFile("%s", "!"); readFile("%s")`, path, path, path)
	val := mustEval(t, input)
	assertString(t, val, "hi!")
}

func TestEvalListDir(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.ToSlash(filepath.Join(dir, "a.txt"))
	pathB := filepath.ToSlash(filepath.Join(dir, "b.txt"))
	input := fmt.Sprintf(`writeFile("%s", "a"); writeFile("%s", "b"); listDir("%s")`, pathA, pathB, filepath.ToSlash(dir))
	val := mustEval(t, input)
	expected := &Array{Elements: []Value{
		&String{Value: "a.txt"},
		&String{Value: "b.txt"},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalTaskThen(t *testing.T) {
	input := `
let work = () -> 3
let task = & work()
let next = task.then(v -> v + 1)
wait next
`
	val := mustEval(t, input)
	assertInteger(t, val, 4)
}

func TestEvalRecoverTaskErrorWhenAllowRecoverableTasks(t *testing.T) {
	input := `
let boom = () -> {
	let obj = {}
	obj.missing
}
let t = & boom()
let out = (wait t) ? { "fallback" }
out
`
	val := mustEval(t, input)
	assertString(t, val, "fallback")
}

func TestEvalRecoverThenTaskErrorWhenAllowRecoverableTasks(t *testing.T) {
	input := `
let work = () -> 1
let t = & work()
let next = t.then(v -> {
	let obj = {}
	obj.missing
})
let out = (wait next) ? { "fallback" }
out
`
	val := mustEval(t, input)
	assertString(t, val, "fallback")
}

func TestEvalRaceExpression(t *testing.T) {
	input := `
let slow = () -> { sleep(10); 1 }
let fast = () -> 2
wait !& { slow(), fast() }
`
	val := mustEval(t, input)
	assertInteger(t, val, 2)
}

func TestEvalRaceCancelsLosers(t *testing.T) {
	input := `
let state = { hits: 0 }
let slow = () -> { sleep(50); state.hits = state.hits + 1; 1 }
let fast = () -> 2
let first = wait !& { slow(), fast() }
sleep(100);
[first, state.hits]
`
	val := mustEval(t, input)
	expected := &Array{Elements: []Value{
		&Integer{Value: 2},
		&Integer{Value: 0},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalSpawnGroup(t *testing.T) {
	val := mustEval(t, "let a = () -> 1; let b = () -> 2; wait & { a(), b() }")
	expected := &Array{Elements: []Value{
		&Integer{Value: 1},
		&Integer{Value: 2},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalSpawnGroupCancelsOnError(t *testing.T) {
	input := `
let state = { hits: 0 }
let boom = () -> fail("boom")
let slow = () -> { sleep(50); state.hits = state.hits + 1; 1 }
let t = & { boom(), slow() }
let out = (wait t) ? { "fallback" }
sleep(100);
[out, state.hits]
`
	val := mustEval(t, input)
	expected := &Array{Elements: []Value{
		&String{Value: "fallback"},
		&Integer{Value: 0},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalTaskCancel(t *testing.T) {
	input := `
let slow = () -> { sleep(1000); 1 }
let t = & slow()
t.cancel()
let out = (wait t) ? { "canceled" }
out
`
	val := mustEval(t, input)
	assertString(t, val, "canceled")
}

func TestEvalUnhandledTaskFailureFailFast(t *testing.T) {
	input := `
let boom = () -> { let obj = {}; obj.missing }
& boom()
sleep(10)
1
`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse failed: %v", errs)
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(input, "<test>")
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if sig != nil {
		t.Fatalf("unexpected signal: %v", sig)
	}
	if err == nil {
		t.Fatalf("expected fail-fast unhandled task failure, got value=%v", val)
	}
	if _, ok := err.(*interpreter.UnhandledTaskError); !ok {
		t.Fatalf("expected UnhandledTaskError, got %T (%v)", err, err)
	}
}

func TestEvalFailFastInterruptsLongSleep(t *testing.T) {
	input := `
let boom = () -> { sleep(20); let obj = {}; obj.missing }
& boom()
sleep(100000)
1
`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse failed: %v", errs)
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(input, "<test>")
	env := interpreter.NewBaseEnvironment()

	done := make(chan struct{})
	var val Value
	var sig *interpreter.Signal
	var err error

	go func() {
		val, sig, err = eval.Eval(program, env)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected fail-fast to interrupt long sleep")
	}

	if sig != nil {
		t.Fatalf("unexpected signal: %v", sig)
	}
	if err == nil {
		t.Fatalf("expected fail-fast error, got value=%v", val)
	}
	if _, ok := err.(*interpreter.UnhandledTaskError); !ok {
		t.Fatalf("expected UnhandledTaskError, got %T (%v)", err, err)
	}
}

func TestEvalUnhandledTaskFailureDeferred(t *testing.T) {
	input := `
let boom = () -> { let obj = {}; obj.missing }
& boom()
sleep(10)
1
`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse failed: %v", errs)
	}

	eval := interpreter.NewEvaluatorWithSourceAndFilename(input, "<test>")
	if err := eval.SetTaskFailurePolicy(interpreter.TaskFailurePolicyDefer); err != nil {
		t.Fatalf("set policy failed: %v", err)
	}
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if err != nil || sig != nil {
		t.Fatalf("unexpected eval result in defer mode: val=%v sig=%v err=%v", val, sig, err)
	}
	if err := eval.CheckUnhandledTaskFailures(); err == nil {
		t.Fatalf("expected deferred unhandled task failure error")
	}
}

func TestEvalMapMethods(t *testing.T) {
	input := `
let m = map()
m.set("a", 1)
m.set(2, "b")
let v = m.get("a")
let hasA = m.has("a")
let hasMissing = m.has("missing")
let removed = m.delete("a")
let stillHasA = m.has("a")
let keys = m.keys()
let vals = m.values()
let keyHas2 = false
let keyHasA = false
let valHasB = false
for i < keys.length with i = 0 {
    if keys[i] == 2 { keyHas2 = true }
    if keys[i] == "a" { keyHasA = true }
    i++
} then ();
for i < vals.length with i = 0 {
    if vals[i] == "b" { valHasB = true }
    i++
} then ();
[v, hasA, hasMissing, removed, stillHasA, keyHas2, keyHasA, keys.length, valHasB, vals.length, m.get(2)]
`
	val := mustEval(t, input)
	expected := &Array{Elements: []Value{
		&Integer{Value: 1},
		&Boolean{Value: true},
		&Boolean{Value: false},
		&Boolean{Value: true},
		&Boolean{Value: false},
		&Boolean{Value: true},
		&Boolean{Value: false},
		&Integer{Value: 1},
		&Boolean{Value: true},
		&Integer{Value: 1},
		&String{Value: "b"},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalImportLocal(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "util.k")
	if err := os.WriteFile(modulePath, []byte(`let answer = 42`), 0o644); err != nil {
		t.Fatalf("write module: %v", err)
	}

	input := `let makeUtil = import "util.k"; let util = makeUtil(); util.answer`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.Fatalf("parse failed")
	}

	eval := interpreter.NewEvaluatorWithSourceFilenameAndRoot(input, "<stdin>", dir)
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if sig != nil {
		t.Fatalf("unexpected signal: %v", sig.Type)
	}
	assertInteger(t, val, 42)
}

func TestEvalImportRelativeToImporterFile(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Import should resolve relative to the importing file's directory, not CWD.
	modulePath := filepath.Join(sub, "util.k")
	if err := os.WriteFile(modulePath, []byte(`let answer = 42`), 0o644); err != nil {
		t.Fatalf("write module: %v", err)
	}

	input := `let makeUtil = import "./util.k"; let util = makeUtil(); util.answer`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.Fatalf("parse failed")
	}

	eval := interpreter.NewEvaluatorWithSourceFilenameAndRoot(input, filepath.Join(sub, "main.k"), dir)
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if sig != nil {
		t.Fatalf("unexpected signal: %v", sig.Type)
	}
	assertInteger(t, val, 42)
}

func TestEvalImportFactoryInstances(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "counter.k")
	module := `
let count = 0
let inc = () -> { count += 1; count }
let get = () -> count
`
	if err := os.WriteFile(modulePath, []byte(module), 0o644); err != nil {
		t.Fatalf("write module: %v", err)
	}

	input := `
let makeCounter = import "counter.k"
let a = makeCounter()
let b = makeCounter()
a.inc()
a.inc()
b.inc()
let output = [a.get(), b.get()]
output
`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.Fatalf("parse failed")
	}

	eval := interpreter.NewEvaluatorWithSourceFilenameAndRoot(input, "<stdin>", dir)
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if sig != nil {
		t.Fatalf("unexpected signal: %v", sig.Type)
	}
	expected := &Array{Elements: []Value{
		&Integer{Value: 2},
		&Integer{Value: 1},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalImportLiveExports(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "counter.k")
	module := `
let count = 0
let get = () -> count
`
	if err := os.WriteFile(modulePath, []byte(module), 0o644); err != nil {
		t.Fatalf("write module: %v", err)
	}

	input := `
let makeCounter = import "counter.k"
let a = makeCounter()
a.count = 4
a.newProp = 7
a.get = () -> 99
let output = [a.count, a.get(), a.newProp]
output
`
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.Fatalf("parse failed")
	}

	eval := interpreter.NewEvaluatorWithSourceFilenameAndRoot(input, "<stdin>", dir)
	env := interpreter.NewBaseEnvironment()
	val, sig, err := eval.Eval(program, env)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if sig != nil {
		t.Fatalf("unexpected signal: %v", sig.Type)
	}
	expected := &Array{Elements: []Value{
		&Integer{Value: 4},
		&Integer{Value: 99},
		&Integer{Value: 7},
	}}
	assertEquivalent(t, val, expected)
}

func TestEvalStringMethods(t *testing.T) {
	input := `
let s = " Hello "
let trimmed = s.trim()
let lower = trimmed.toLower()
let upper = trimmed.toUpper()
let parts = trimmed.split("l")
let chars = trimmed.chars()
let has = trimmed.contains("ell")
let starts = trimmed.startsWith("He")
let ends = trimmed.endsWith("lo")
let replaced = trimmed.replace("l", "x");
[trimmed, lower, upper, parts.length, chars.length, has, starts, ends, replaced]
`
	val := mustEval(t, input)
	expected := &Array{Elements: []Value{
		&String{Value: "Hello"},
		&String{Value: "hello"},
		&String{Value: "HELLO"},
		&Integer{Value: 3},
		&Integer{Value: 5},
		&Boolean{Value: true},
		&Boolean{Value: true},
		&Boolean{Value: true},
		&String{Value: "Hexxo"},
	}}
	assertEquivalent(t, val, expected)
}
