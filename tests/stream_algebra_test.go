package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestStreamMapFilterFlatMapCollect(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "in.log")
	content := "a\nbb\nccc\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
read(%q)
| lines()
| map(x -> x.toUpper())
| filter(x -> x.length >= 2)
| flatMap(x -> [x, x + "!"])
| collect()
`, inPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(arr.Elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(arr.Elements))
	}
	assertString(t, arr.Elements[0], "BB")
	assertString(t, arr.Elements[1], "BB!")
	assertString(t, arr.Elements[2], "CCC")
	assertString(t, arr.Elements[3], "CCC!")
}

func TestStreamCountSink(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "count.log")
	content := "alpha\nbeta\ngamma\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`read(%q) | lines() | count()`, inPath)
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertInteger(t, val, 3)
}

func TestStreamReduceSink(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "reduce.log")
	content := "a\nb\nc\n"
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
read(%q)
| lines()
| reduce("", (acc, x) -> acc + "[" + x + "]")
`, inPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertString(t, val, "[a][b][c]")
}

func TestStreamFilterRequiresBoolPredicate(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "filter.log")
	if err := os.WriteFile(inPath, []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`read(%q) | lines() | filter(x -> x) | collect()`, inPath)
	_, err := evalInput(t, input)
	if err == nil {
		t.Fatalf("expected runtime error")
	}
	if !strings.Contains(err.Error(), "filter predicate must return bool") {
		t.Fatalf("unexpected runtime error: %v", err)
	}
}

func TestStreamFlatMapRequiresArray(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "flatmap.log")
	if err := os.WriteFile(inPath, []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`read(%q) | lines() | flatMap(x -> x) | collect()`, inPath)
	_, err := evalInput(t, input)
	if err == nil {
		t.Fatalf("expected runtime error")
	}
	if !strings.Contains(err.Error(), "flatMap mapper must return array") {
		t.Fatalf("unexpected runtime error: %v", err)
	}
}

func TestStreamTakeDropStages(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "take-drop.log")
	if err := os.WriteFile(inPath, []byte("a\nb\nc\nd\ne\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`read(%q) | lines() | drop(1) | take(2) | collect()`, inPath)
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
	}
	assertString(t, arr.Elements[0], "b")
	assertString(t, arr.Elements[1], "c")
}

func TestStreamChunkStage(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "chunk.log")
	if err := os.WriteFile(inPath, []byte("a\nb\nc\nd\ne\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`read(%q) | lines() | chunk(2) | collect()`, inPath)
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	out, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(out.Elements) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(out.Elements))
	}
	assertChunk(t, out.Elements[0], []string{"a", "b"})
	assertChunk(t, out.Elements[1], []string{"c", "d"})
	assertChunk(t, out.Elements[2], []string{"e"})
}

func TestStreamWindowStage(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "window.log")
	if err := os.WriteFile(inPath, []byte("a\nb\nc\nd\ne\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`read(%q) | lines() | window(3, 2) | collect()`, inPath)
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	out, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(out.Elements) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(out.Elements))
	}
	assertChunk(t, out.Elements[0], []string{"a", "b", "c"})
	assertChunk(t, out.Elements[1], []string{"c", "d", "e"})
}

func assertChunk(t *testing.T, val Value, expected []string) {
	t.Helper()
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected chunk array, got %T", val)
	}
	if len(arr.Elements) != len(expected) {
		t.Fatalf("expected chunk length %d, got %d", len(expected), len(arr.Elements))
	}
	for i, item := range expected {
		assertString(t, arr.Elements[i], item)
	}
}

func TestStreamSendSinkToChannel(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "send.log")
	if err := os.WriteFile(inPath, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let c = buffered(8)
let producer = & (read(%q) | lines() | send(c))
let out = for true with acc = [] {
    let [v, done] = c.recv()
    if done { break acc }
    acc += [v]
} then ()
wait producer
out
`, inPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 values, got %d", len(arr.Elements))
	}
	assertString(t, arr.Elements[0], "a")
	assertString(t, arr.Elements[1], "b")
	assertString(t, arr.Elements[2], "c")
}

func TestStreamFromChannelSourceCollect(t *testing.T) {
	input := `
let c = channel()
let producer = & (() -> {
    c.send("x")
    c.send("y")
    c.done()
})()
let out = fromChannel(c) | collect()
wait producer
out
`

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 values, got %d", len(arr.Elements))
	}
	assertString(t, arr.Elements[0], "x")
	assertString(t, arr.Elements[1], "y")
}

func TestStreamMergeSource(t *testing.T) {
	tempDir := t.TempDir()
	aPath := filepath.Join(tempDir, "a.log")
	bPath := filepath.Join(tempDir, "b.log")
	if err := os.WriteFile(aPath, []byte("a1\na2\n"), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(bPath, []byte("b1\nb2\n"), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}

	input := fmt.Sprintf(`
merge(
  read(%q) | lines(),
  read(%q) | lines(),
) | collect()
`, aPath, bPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(arr.Elements) != 4 {
		t.Fatalf("expected 4 values, got %d", len(arr.Elements))
	}
	out := []string{
		arr.Elements[0].(*String).Value,
		arr.Elements[1].(*String).Value,
		arr.Elements[2].(*String).Value,
		arr.Elements[3].(*String).Value,
	}
	sort.Strings(out)
	expected := []string{"a1", "a2", "b1", "b2"}
	for i := range expected {
		if out[i] != expected[i] {
			t.Fatalf("unexpected merged value at %d: got %q want %q", i, out[i], expected[i])
		}
	}
}

func TestStreamZipSource(t *testing.T) {
	tempDir := t.TempDir()
	leftPath := filepath.Join(tempDir, "left.log")
	rightPath := filepath.Join(tempDir, "right.log")
	if err := os.WriteFile(leftPath, []byte("L1\nL2\nL3\n"), 0o644); err != nil {
		t.Fatalf("write left: %v", err)
	}
	if err := os.WriteFile(rightPath, []byte("R1\nR2\n"), 0o644); err != nil {
		t.Fatalf("write right: %v", err)
	}

	input := fmt.Sprintf(`
zip(
  read(%q) | lines(),
  read(%q) | lines(),
) | collect()
`, leftPath, rightPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	out, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(out.Elements) != 2 {
		t.Fatalf("expected 2 zipped pairs, got %d", len(out.Elements))
	}
	assertChunk(t, out.Elements[0], []string{"L1", "R1"})
	assertChunk(t, out.Elements[1], []string{"L2", "R2"})
}

func TestStreamPartitionPassBranch(t *testing.T) {
	input := `
let c = rendezvous()
let p = fromChannel(c) | partition(x -> x.length >= 2)

let passTask = & (() -> p.pass | collect())()

sleep(5)
c.send("a")
c.send("bb")
c.send("ccc")
c.done()

wait passTask
`
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	pass, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected pass array, got %T", val)
	}
	if len(pass.Elements) != 2 {
		t.Fatalf("unexpected pass size: %d", len(pass.Elements))
	}
	assertString(t, pass.Elements[0], "bb")
	assertString(t, pass.Elements[1], "ccc")
}

func TestStreamPartitionFailBranch(t *testing.T) {
	input := `
let c = rendezvous()
let p = fromChannel(c) | partition(x -> x.length >= 2)

let failTask = & (() -> p.fail | collect())()

sleep(5)
c.send("a")
c.send("bb")
c.send("ccc")
c.done()

wait failTask
`
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	fail, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected fail array, got %T", val)
	}
	if len(fail.Elements) != 1 {
		t.Fatalf("unexpected fail size: %d", len(fail.Elements))
	}
	assertString(t, fail.Elements[0], "a")
}

func TestStreamPartitionKeyedErrBranch(t *testing.T) {
	input := `
let c = rendezvous()
let p = fromChannel(c) | partition(
    x -> if x.startsWith("err") { "err" } else if x.startsWith("warn") { "warn" } else { "info" },
    ["err", "warn", "info"],
)

let errTask = & (() -> p.err | collect())()

sleep(5)
c.send("info:ok")
c.send("warn:cpu")
c.send("err:disk")
c.send("debug:skip")
c.send("warn:mem")
c.done()

wait errTask
`
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	errRows, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected err array, got %T", val)
	}
	if len(errRows.Elements) != 1 {
		t.Fatalf("unexpected err size: %d", len(errRows.Elements))
	}
	assertString(t, errRows.Elements[0], "err:disk")
}

func TestStreamPartitionDropsUnconsumedBranches(t *testing.T) {
	input := `
let c = rendezvous()
let p = fromChannel(c) | partition(
    x -> if x.startsWith("ok") { "ok" } else { "drop" },
    ["ok", "drop"],
)

let okTask = & (() -> p.ok | collect())()

sleep(50)
c.send("ok-1")
c.send("drop-1")
c.send("ok-2")
c.done()

wait okTask
`
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	okRows, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected ok array, got %T", val)
	}
	if len(okRows.Elements) != 2 {
		t.Fatalf("expected 2 kept values, got %d", len(okRows.Elements))
	}
	assertString(t, okRows.Elements[0], "ok-1")
	assertString(t, okRows.Elements[1], "ok-2")
}

func TestStreamThrottleStage(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "throttle.log")
	if err := os.WriteFile(inPath, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`read(%q) | lines() | throttle(1) | count()`, inPath)
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertInteger(t, val, 3)
}

func TestStreamToChannelSink(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "tochannel.log")
	if err := os.WriteFile(inPath, []byte("x\ny\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let c = buffered(8)
let t = & (read(%q) | lines() | toChannel(c))
let out = for true with acc = [] {
    let [v, done] = c.recv()
    if done { break acc }
    acc += [v]
} then ()
wait t
out
`, inPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 values, got %d", len(arr.Elements))
	}
	assertString(t, arr.Elements[0], "x")
	assertString(t, arr.Elements[1], "y")
}

func TestStreamDebounceStage(t *testing.T) {
	input := `
let c = buffered(32)
c.send("a")
c.send("b")
c.send("c")
c.done()
fromChannel(c) | debounce(5) | collect()
`
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(arr.Elements) != 1 {
		t.Fatalf("expected 1 debounced item, got %d", len(arr.Elements))
	}
	assertString(t, arr.Elements[0], "c")
}

func TestStreamJoinSourceByKey(t *testing.T) {
	tempDir := t.TempDir()
	leftPath := filepath.Join(tempDir, "left-join.log")
	rightPath := filepath.Join(tempDir, "right-join.log")
	if err := os.WriteFile(leftPath, []byte("u1|Alice\nu2|Bob\nu3|Carol\n"), 0o644); err != nil {
		t.Fatalf("write left: %v", err)
	}
	if err := os.WriteFile(rightPath, []byte("u2|active\nu1|trial\nu2|pro\n"), 0o644); err != nil {
		t.Fatalf("write right: %v", err)
	}

	input := fmt.Sprintf(`
let left = read(%q) | lines() | map(line -> {
    let p = line.split("|")
    { id: p[0], name: p[1], }
})
let right = read(%q) | lines() | map(line -> {
    let p = line.split("|")
    { id: p[0], plan: p[1], }
})

join(left, right, l -> l.id, r -> r.id)
    | map(pair -> { leftId: pair[0].id, name: pair[0].name, plan: pair[1].plan, })
    | collect()
`, leftPath, rightPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	out, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(out.Elements) != 3 {
		t.Fatalf("expected 3 joined rows, got %d", len(out.Elements))
	}
	first, _ := out.Elements[0].(*Object)
	second, _ := out.Elements[1].(*Object)
	third, _ := out.Elements[2].(*Object)
	assertString(t, first.Pairs["leftId"], "u1")
	assertString(t, first.Pairs["name"], "Alice")
	assertString(t, first.Pairs["plan"], "trial")
	assertString(t, second.Pairs["leftId"], "u2")
	assertString(t, second.Pairs["name"], "Bob")
	assertString(t, second.Pairs["plan"], "active")
	assertString(t, third.Pairs["leftId"], "u2")
	assertString(t, third.Pairs["name"], "Bob")
	assertString(t, third.Pairs["plan"], "pro")
}

func TestStreamImplicitLambdaShorthand(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "implicit.log")
	if err := os.WriteFile(inPath, []byte("a\nbb\nccc\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`read(%q) | lines() | filter(_ != "") | map("x:" + _) | collect()`, inPath)
	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	arr, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 values, got %d", len(arr.Elements))
	}
	assertString(t, arr.Elements[0], "x:a")
	assertString(t, arr.Elements[1], "x:bb")
	assertString(t, arr.Elements[2], "x:ccc")
}

func TestStreamTeeStage(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "tee-in.log")
	outPath := filepath.Join(tempDir, "tee-out.log")
	if err := os.WriteFile(inPath, []byte("a\nb\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let side = write(%q, { type: TEXT, })
let rows = read(%q) | lines() | tee(side) | collect()
;
[rows, readFile(%q)]
`, outPath, inPath, outPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	out, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	rows, ok := out.Elements[0].(*Array)
	if !ok {
		t.Fatalf("expected rows array, got %T", out.Elements[0])
	}
	if len(rows.Elements) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows.Elements))
	}
	assertString(t, rows.Elements[0], "a")
	assertString(t, rows.Elements[1], "b")
	assertString(t, out.Elements[1], "a\nb\n")
}

func TestStreamSpillStage(t *testing.T) {
	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "spill-in.log")
	outPath := filepath.Join(tempDir, "spill-out.log")
	if err := os.WriteFile(inPath, []byte("a\nb\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	input := fmt.Sprintf(`
let n = read(%q) | lines() | spill(%q, { type: TEXT, }) | count()
;
[n, readFile(%q)]
`, inPath, outPath, outPath)

	val, err := evalInput(t, input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	out, ok := val.(*Array)
	if !ok {
		t.Fatalf("expected array, got %T", val)
	}
	assertInteger(t, out.Elements[0], 2)
	assertString(t, out.Elements[1], "a\nb\n")
}
