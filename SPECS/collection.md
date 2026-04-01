# Collections Specification

Status: draft  
Scope: collection data structures and their runtime contracts.

This spec defines Karl collections as first-class runtime types with predictable
complexity and explicit semantics.

## 1. Design Goals

Karl collections should be:

- useful for real programs, not only toy storage
- explicit about asymptotic behavior
- predictable across CLI/REPL/bench
- composable with Karl expression style

## 2. Collection Families

Karl has two different tree families with different purposes:

1. Ordered index tree (`tree`) for O(log n) key lookup and bounds.
2. Hierarchical node tree (`ntree`) for parent/children/siblings graph-like navigation.

Do not conflate these two models.

## 3. Existing Collections (Current)

### 3.1 Array

Literal:

```karl
let xs = [1, 2, 3]
```

Core API (current):

- `xs.length`
- `xs.map(fn)`
- `xs.filter(fn)`
- `xs.reduce(fn, init)`
- `xs.find(fn)`
- `xs.sort(cmp)`
- `xs.sum()`
- `xs.push(v)`

### 3.2 Map

Constructor:

```karl
let m = map()
```

Core API (current):

- `m.set(key, value)`
- `m.get(key) -> value|null`
- `m.has(key) -> bool`
- `m.delete(key) -> bool`
- `m.keys() -> []`
- `m.values() -> []`

### 3.3 Set

Constructor:

```karl
let s = set()
```

Core API (current):

- `s.add(v)`
- `s.has(v) -> bool`
- `s.delete(v) -> bool`
- `s.values() -> []`
- `s.size`

## 4. Ordered Index Tree: `tree(kind?)` (Current)

Purpose: ordered key-value index with logarithmic operations.

Constructor:

```karl
let idx = tree()         // default: "avl"
let fast = tree("treap")
```

Key constraints (current):

- keys must be homogeneous and ordered scalar values: `int | float | string | char`
- mixed key domains are runtime errors

Current API:

- `idx.set(key, value) -> tree`
- `idx.get(key) -> value|null`
- `idx.has(key) -> bool`
- `idx.delete(key) -> bool`
- `idx.kind() -> string`
- `idx.min() -> { key, value } | null`
- `idx.max() -> { key, value } | null`
- `idx.lowerBound(key) -> { key, value } | null` (`>= key`)
- `idx.upperBound(key) -> { key, value } | null` (`> key`)
- `idx.floor(key) -> { key, value } | null` (`<= key`)
- `idx.ceil(key) -> { key, value } | null` (`>= key`)
- `idx.predecessor(key) -> { key, value } | null` (`< key`)
- `idx.successor(key) -> { key, value } | null` (`> key`)
- `idx.closest(key, opts?) -> { key, value, exact } | null`
  - exact match: `exact=true`
  - tie behavior is deterministic:
    - default: choose lower key on equal distance
    - optional: `opts.tie = "lower" | "upper"`
  - numeric distance is defined for `int`/`float`
  - for non-numeric key domains, `closest` raises runtime error
- `idx.range(from, to, opts?) -> [{ key, value }]`
  - `opts.includeFrom` default `true`
  - `opts.includeTo` default `true`
  - `opts.limit` optional positive int
- `idx.keys() -> []`
- `idx.values() -> []`
- `idx.items() -> [{ key, value }]`
- `idx.size`

Inspect rendering:

- empty tree: `tree(<kind>, size=0)`
- non-empty tree: header + ASCII branch rendering (Unix `tree` style)

Complexity targets:

- `set/get/has/delete/floor/ceil/predecessor/successor/closest`: O(log n)
- `keys/values/items/range`: O(k + log n), where `k` is output size

## 5. Hierarchical Node Tree: `ntree(...)` (Current)

Purpose: mutable hierarchy with stable node identity and relationship navigation.

This is the API for parent/children/siblings use cases.

### 5.1 Why a Separate Type

`tree("avl"|"treap")` rebalances internally. Parent/child relationships are
implementation artifacts and may change after inserts/deletes. Exposing those
relationships in ordered trees would be unstable and misleading.

A dedicated hierarchical tree keeps parent/children as user-level semantics.

### 5.2 Constructor and Node Model

Constructor:

```karl
let t = ntree("root", { label: "Root", })
```

Node shape (conceptual):

```karl
{ id, parent, value, children, }
```

Where:

- `id`: unique node identifier (`string`)
- `parent`: parent id or `null` for root
- `value`: arbitrary Karl value
- `children`: ordered array of child ids

### 5.3 Hierarchical API

- `t.get(id) -> node|null`
- `t.set(id, value) -> bool`
- `t.append(parentId, childId, value) -> ntree`
- `t.prepend(parentId, childId, value) -> ntree`
- `t.insertAt(parentId, index, childId, value) -> ntree`
- `t.remove(id, opts?) -> bool`
  - default `opts.subtree=true` (remove full subtree)
- `t.move(id, newParentId, opts?) -> ntree`
  - optional insertion index
- `t.parent(id) -> node|null`
- `t.children(id) -> [node]`
- `t.siblings(id, opts?) -> [node]`
  - default excludes self
- `t.ancestors(id) -> [node]` (nearest first)
- `t.descendants(id, opts?) -> [node]`
  - traversal option: `dfs` (default) or `bfs`
- `t.find(fn) -> node|null`
- `t.findAll(fn) -> [node]`
- `t.root() -> node`
- `t.size`

Inspect rendering:

- `log(t)` prints Unix tree-like hierarchy using node ids/labels

### 5.4 Complexity Targets

With internal `nodesById`, `parentById`, `childrenById` maps:

- `get/parent/children`: O(1) (plus O(k) to materialize children list)
- `append/prepend`: O(1) amortized
- `insertAt/remove/move`: O(c) where `c` is affected sibling list size
- `find/findAll/descendants`: O(n)

## 6. Error Semantics

Collection operations follow Karl runtime error model:

- invalid argument type -> runtime error
- invalid key domain transition in ordered tree -> runtime error
- missing node/key where API expects optional value -> `null`/`false`, not panic

No host-language panics must escape.

## 7. Naming Rules

Collection APIs should use clear semantics:

- ordered-tree search vocabulary: `floor`, `ceil`, `predecessor`, `successor`, `closest`
- hierarchy vocabulary: `parent`, `children`, `siblings`, `ancestors`, `descendants`

Avoid aliases unless explicitly approved.

## 8. Non-Goals

- exposing internal balancing structure of ordered trees as stable user semantics
- implicit auto-indexing for `find` in hierarchical trees
- automatic persistence/serialization in collection core API

## 9. Compatibility Notes

Current programs using `tree` remain valid.  
`ntree` is additive and does not replace ordered `tree`.
