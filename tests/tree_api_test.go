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
  lb: t.lowerBound(1),
  ub: t.upperBound(1),
  items: t.items(),
  keys: t.keys(),
  values: t.values(),
}
`)
	expected := &Object{Pairs: map[string]Value{
		"min":    NullValue,
		"max":    NullValue,
		"lb":     NullValue,
		"ub":     NullValue,
		"items":  &Array{Elements: []Value{}},
		"keys":   &Array{Elements: []Value{}},
		"values": &Array{Elements: []Value{}},
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
