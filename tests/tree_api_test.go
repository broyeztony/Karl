package tests

import (
	"strings"
	"testing"
)

func TestTreeAVLBasicOps(t *testing.T) {
	val := mustEval(t, `
let t = tree()
t.set(5, "e")
t.set(2, "b")
t.set(8, "h")
t.set(3, "c")

	{
	  kind: t.kind(),
	  get2: t.get(2),
	  has7: t.has(7),
	  has8: t.has(8),
	  min: t.min(),
	  max: t.max(),
	  maxDepth: t.maxDepth(),
	  maxWidth: t.maxWidth(),
	  lb4: t.lowerBound(4),
	  ub5: t.upperBound(5),
	  keys: t.keys(),
	  vals: t.values(),
	  size: t.size,
  lenv: len(t),
}
`)
	expected := &Object{Pairs: map[string]Value{
		"kind": &String{Value: "avl"},
		"get2": &String{Value: "b"},
		"has7": &Boolean{Value: false},
		"has8": &Boolean{Value: true},
		"min": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 2},
			"value": &String{Value: "b"},
		}},
		"max": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 8},
			"value": &String{Value: "h"},
		}},
		"maxDepth": &Integer{Value: 3},
		"maxWidth": &Integer{Value: 2},
		"lb4": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 5},
			"value": &String{Value: "e"},
		}},
		"ub5": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 8},
			"value": &String{Value: "h"},
		}},
		"keys": &Array{Elements: []Value{
			&Integer{Value: 2},
			&Integer{Value: 3},
			&Integer{Value: 5},
			&Integer{Value: 8},
		}},
		"vals": &Array{Elements: []Value{
			&String{Value: "b"},
			&String{Value: "c"},
			&String{Value: "e"},
			&String{Value: "h"},
		}},
		"size": &Integer{Value: 4},
		"lenv": &Integer{Value: 4},
	}}
	assertEquivalent(t, val, expected)
}

func TestTreeTreapBasicOps(t *testing.T) {
	val := mustEval(t, `
let t = tree("treap")
t.set("d", 4)
t.set("a", 1)
t.set("c", 3)
t.set("b", 2)
t.delete("c")
{
  kind: t.kind(),
  keys: t.keys(),
  values: t.values(),
  hasC: t.has("c"),
  lbBb: t.lowerBound("bb"),
}
`)
	expected := &Object{Pairs: map[string]Value{
		"kind": &String{Value: "treap"},
		"keys": &Array{Elements: []Value{
			&String{Value: "a"},
			&String{Value: "b"},
			&String{Value: "d"},
		}},
		"values": &Array{Elements: []Value{
			&Integer{Value: 1},
			&Integer{Value: 2},
			&Integer{Value: 4},
		}},
		"hasC": &Boolean{Value: false},
		"lbBb": &Object{Pairs: map[string]Value{
			"key":   &String{Value: "d"},
			"value": &Integer{Value: 4},
		}},
	}}
	assertEquivalent(t, val, expected)
}

func TestTreeRejectsMixedKeyTypes(t *testing.T) {
	_, err := evalInput(t, `
let t = tree()
t.set(1, "a")
t.set("x", "b")
`)
	if err == nil {
		t.Fatalf("expected mixed key type error")
	}
	if !strings.Contains(err.Error(), "tree key type mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTreeInvalidKind(t *testing.T) {
	_, err := evalInput(t, `tree("btree")`)
	if err == nil {
		t.Fatalf("expected invalid tree kind error")
	}
	if !strings.Contains(err.Error(), "tree kind must be") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTreeItemMethodsOnEmpty(t *testing.T) {
	val := mustEval(t, `
let t = tree()
	{
	  min: t.min(),
	  max: t.max(),
	  maxDepth: t.maxDepth(),
	  maxWidth: t.maxWidth(),
	  lb: t.lowerBound(1),
	  ub: t.upperBound(1),
	  items: t.items(),
	  keys: t.keys(),
	  values: t.values(),
}
`)
	expected := &Object{Pairs: map[string]Value{
		"min":      NullValue,
		"max":      NullValue,
		"maxDepth": &Integer{Value: 0},
		"maxWidth": &Integer{Value: 0},
		"lb":       NullValue,
		"ub":       NullValue,
		"items":    &Array{Elements: []Value{}},
		"keys":     &Array{Elements: []Value{}},
		"values":   &Array{Elements: []Value{}},
	}}
	assertEquivalent(t, val, expected)
}

func TestTreeInspectPrintsAsciiTree(t *testing.T) {
	val := mustEval(t, `
let t = tree()
t.set(5, "e")
t.set(2, "b")
t.set(8, "h")
t.set(3, "c")
t
`)
	got := val.Inspect()
	want := "tree(avl, size=4)\n`-- 5: \"e\"\n    |-- 2: \"b\"\n    |   `-- 3: \"c\"\n    `-- 8: \"h\""
	if got != want {
		t.Fatalf("unexpected inspect output:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestTreeOrderedSearchExtensions(t *testing.T) {
	val := mustEval(t, `
let t = tree()
t.set(10, "a")
t.set(20, "b")
t.set(30, "c")
t.set(40, "d")
{
  floor25: t.floor(25),
  ceil25: t.ceil(25),
  pred30: t.predecessor(30),
  succ30: t.successor(30),
  closest25: t.closest(25),
  closest25Up: t.closest(25, { tie: "upper", }),
  closest30: t.closest(30),
}
`)
	expected := &Object{Pairs: map[string]Value{
		"floor25": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 20},
			"value": &String{Value: "b"},
		}},
		"ceil25": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 30},
			"value": &String{Value: "c"},
		}},
		"pred30": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 20},
			"value": &String{Value: "b"},
		}},
		"succ30": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 40},
			"value": &String{Value: "d"},
		}},
		"closest25": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 20},
			"value": &String{Value: "b"},
			"exact": &Boolean{Value: false},
		}},
		"closest25Up": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 30},
			"value": &String{Value: "c"},
			"exact": &Boolean{Value: false},
		}},
		"closest30": &Object{Pairs: map[string]Value{
			"key":   &Integer{Value: 30},
			"value": &String{Value: "c"},
			"exact": &Boolean{Value: true},
		}},
	}}
	assertEquivalent(t, val, expected)
}

func TestTreeRangeOptions(t *testing.T) {
	val := mustEval(t, `
let t = tree()
t.set(10, "a")
t.set(20, "b")
t.set(30, "c")
t.set(40, "d")
{
  inc: t.range(20, 40),
  exc: t.range(20, 40, { includeFrom: false, includeTo: false, }),
  lim: t.range(10, 40, { limit: 2, }),
}
`)
	expected := &Object{Pairs: map[string]Value{
		"inc": &Array{Elements: []Value{
			&Object{Pairs: map[string]Value{"key": &Integer{Value: 20}, "value": &String{Value: "b"}}},
			&Object{Pairs: map[string]Value{"key": &Integer{Value: 30}, "value": &String{Value: "c"}}},
			&Object{Pairs: map[string]Value{"key": &Integer{Value: 40}, "value": &String{Value: "d"}}},
		}},
		"exc": &Array{Elements: []Value{
			&Object{Pairs: map[string]Value{"key": &Integer{Value: 30}, "value": &String{Value: "c"}}},
		}},
		"lim": &Array{Elements: []Value{
			&Object{Pairs: map[string]Value{"key": &Integer{Value: 10}, "value": &String{Value: "a"}}},
			&Object{Pairs: map[string]Value{"key": &Integer{Value: 20}, "value": &String{Value: "b"}}},
		}},
	}}
	assertEquivalent(t, val, expected)
}

func TestTreeRangeLimitStrictRegression(t *testing.T) {
	val := mustEval(t, `
let t = tree()
let seeds = [10, 20, 35, 50, 80]
seeds.forEach(k -> t.set(k, "v" + str(k)))
t.range(10, 80, { limit: 3, }).length
`)
	assertInteger(t, val, 3)
}

func TestTreeClosestRejectsNonNumericKeys(t *testing.T) {
	_, err := evalInput(t, `
let t = tree()
t.set("a", 1)
t.closest("b")
`)
	if err == nil {
		t.Fatalf("expected closest key type error")
	}
	if !strings.Contains(err.Error(), "closest expects numeric tree keys") {
		t.Fatalf("unexpected error: %v", err)
	}
}
