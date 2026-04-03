package tests

import "testing"

func TestNTreeBasicOps(t *testing.T) {
	val := mustEval(t, `
let t = ntree("root", { label: "Root", })
t.append("root", "a", { v: 1, })
t.prepend("root", "b", { v: 2, })
t.insertAt("root", 1, "c", { v: 3, })
t.append("a", "a1", { v: 11, })
	{
	  size: t.size,
	  lenv: len(t),
	  root: t.root().id,
	  getA: t.get("a").id,
	  parentA1: t.parent("a1").id,
	  pathA1: t.path("a1").map(n -> n.id),
	  pathMissing: t.path("zzz"),
	  childrenRoot: t.children("root").map(n -> n.id),
	  siblingsC: t.siblings("c").map(n -> n.id),
	  ancA1: t.ancestors("a1").map(n -> n.id),
  descRootDfs: t.descendants("root").map(n -> n.id),
  descRootBfs: t.descendants("root", { traversal: "bfs", }).map(n -> n.id),
  findC: t.find(n -> n.id == "c").id,
  findAllA: t.findAll(n -> n.id.startsWith("a")).map(n -> n.id),
}
`)

	expected := &Object{Pairs: map[string]Value{
		"size":     &Integer{Value: 5},
		"lenv":     &Integer{Value: 5},
		"root":     &String{Value: "root"},
		"getA":     &String{Value: "a"},
		"parentA1": &String{Value: "a"},
		"pathA1": &Array{Elements: []Value{
			&String{Value: "root"},
			&String{Value: "a"},
			&String{Value: "a1"},
		}},
		"pathMissing": NullValue,
		"childrenRoot": &Array{Elements: []Value{
			&String{Value: "b"},
			&String{Value: "c"},
			&String{Value: "a"},
		}},
		"siblingsC": &Array{Elements: []Value{
			&String{Value: "b"},
			&String{Value: "a"},
		}},
		"ancA1": &Array{Elements: []Value{
			&String{Value: "a"},
			&String{Value: "root"},
		}},
		"descRootDfs": &Array{Elements: []Value{
			&String{Value: "b"},
			&String{Value: "c"},
			&String{Value: "a"},
			&String{Value: "a1"},
		}},
		"descRootBfs": &Array{Elements: []Value{
			&String{Value: "b"},
			&String{Value: "c"},
			&String{Value: "a"},
			&String{Value: "a1"},
		}},
		"findC": &String{Value: "c"},
		"findAllA": &Array{Elements: []Value{
			&String{Value: "a"},
			&String{Value: "a1"},
		}},
	}}
	assertEquivalent(t, val, expected)
}

func TestNTreeMoveAndRemove(t *testing.T) {
	val := mustEval(t, `
let t = ntree("root", 0)
t.append("root", "a", 1)
t.append("root", "b", 2)
t.append("root", "c", 3)
t.append("a", "a1", 11)
t.append("a", "a2", 12)
t.move("c", "a", { index: 1, })
t.remove("a", { subtree: false, })
t.remove("b")
{
  size: t.size,
  childrenRoot: t.children("root").map(n -> n.id),
  parentC: t.parent("c").id,
  hasA: t.get("a") == null,
  hasB: t.get("b") == null,
}
`)

	expected := &Object{Pairs: map[string]Value{
		"size": &Integer{Value: 4},
		"childrenRoot": &Array{Elements: []Value{
			&String{Value: "a1"},
			&String{Value: "c"},
			&String{Value: "a2"},
		}},
		"parentC": &String{Value: "root"},
		"hasA":    &Boolean{Value: true},
		"hasB":    &Boolean{Value: true},
	}}
	assertEquivalent(t, val, expected)
}

func TestNTreeInspectPrintsAsciiTree(t *testing.T) {
	val := mustEval(t, `
let t = ntree("root", "R")
t.append("root", "a", "A")
t.append("root", "b", "B")
t.append("a", "a1", "A1")
t
`)
	want := "ntree(size=4, root=\"root\")\n`-- \"root\": \"R\"\n    |-- \"a\": \"A\"\n    |   `-- \"a1\": \"A1\"\n    `-- \"b\": \"B\""
	if got := val.Inspect(); got != want {
		t.Fatalf("unexpected inspect output:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestNTreeTopologyMethods(t *testing.T) {
	val := mustEval(t, `
let t = ntree("root", 0)
t.append("root", "a", 1)
t.append("root", "b", 2)
t.append("root", "c", 3)
t.append("a", "a1", 11)
t.append("a", "a2", 12)
t.append("a2", "a21", 121)
t.append("c", "c1", 31)
{
  depthRoot: t.depth("root"),
  depthA21: t.depth("a21"),
  depthMissing: t.depth("nope"),
  heightRoot: t.height("root"),
  heightA: t.height("a"),
  heightB: t.height("b"),
  lcaNear: t.lca("a1", "a21").id,
  lcaFar: t.lca("a1", "c1").id,
  lcaMissing: t.lca("a1", "missing"),
  pathNear: t.pathBetween("a1", "a21").map(n -> n.id),
  pathFar: t.pathBetween("a1", "c1").map(n -> n.id),
  pathMissing: t.pathBetween("a1", "missing"),
  subA: t.subtreeSize("a"),
  subRoot: t.subtreeSize("root"),
  subMissing: t.subtreeSize("missing"),
}
`)
	expected := &Object{Pairs: map[string]Value{
		"depthRoot":    &Integer{Value: 0},
		"depthA21":     &Integer{Value: 3},
		"depthMissing": NullValue,
		"heightRoot":   &Integer{Value: 3},
		"heightA":      &Integer{Value: 2},
		"heightB":      &Integer{Value: 0},
		"lcaNear":      &String{Value: "a"},
		"lcaFar":       &String{Value: "root"},
		"lcaMissing":   NullValue,
		"pathNear": &Array{Elements: []Value{
			&String{Value: "a1"},
			&String{Value: "a"},
			&String{Value: "a2"},
			&String{Value: "a21"},
		}},
		"pathFar": &Array{Elements: []Value{
			&String{Value: "a1"},
			&String{Value: "a"},
			&String{Value: "root"},
			&String{Value: "c"},
			&String{Value: "c1"},
		}},
		"pathMissing": NullValue,
		"subA":        &Integer{Value: 4},
		"subRoot":     &Integer{Value: 8},
		"subMissing":  NullValue,
	}}
	assertEquivalent(t, val, expected)
}
