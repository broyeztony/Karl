package interpreter

import (
	"math"
	"math/rand"
	"sort"
	"strings"
)

type treeKey struct {
	Type ValueType
	I    int64
	F    float64
	S    string
}

type treeNode struct {
	key      treeKey
	value    Value
	left     *treeNode
	right    *treeNode
	height   int
	priority int
}

func newTree(kind string) (*Tree, error) {
	k := strings.ToLower(strings.TrimSpace(kind))
	if k == "" {
		k = "avl"
	}
	if k != "avl" && k != "treap" {
		return nil, &RuntimeError{Message: "tree kind must be \"avl\" or \"treap\""}
	}
	return &Tree{kind: k}, nil
}

func treeKeyForValue(v Value) (treeKey, error) {
	switch x := v.(type) {
	case *Integer:
		return treeKey{Type: INTEGER, I: x.Value}, nil
	case *Float:
		return treeKey{Type: FLOAT, F: x.Value}, nil
	case *String:
		return treeKey{Type: STRING, S: x.Value}, nil
	case *Char:
		return treeKey{Type: CHAR, S: x.Value}, nil
	default:
		return treeKey{}, &RuntimeError{Message: "tree keys must be int, float, string, or char"}
	}
}

func treeKeyToValue(k treeKey) Value {
	switch k.Type {
	case INTEGER:
		return &Integer{Value: k.I}
	case FLOAT:
		return &Float{Value: k.F}
	case STRING:
		return &String{Value: k.S}
	case CHAR:
		return &Char{Value: k.S}
	default:
		return NullValue
	}
}

func compareTreeKey(a, b treeKey) (int, error) {
	if a.Type != b.Type {
		return 0, &RuntimeError{Message: "tree key type mismatch"}
	}
	switch a.Type {
	case INTEGER:
		if a.I < b.I {
			return -1, nil
		}
		if a.I > b.I {
			return 1, nil
		}
		return 0, nil
	case FLOAT:
		if a.F < b.F {
			return -1, nil
		}
		if a.F > b.F {
			return 1, nil
		}
		return 0, nil
	case STRING, CHAR:
		if a.S < b.S {
			return -1, nil
		}
		if a.S > b.S {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, &RuntimeError{Message: "unsupported tree key type"}
	}
}

func (t *Tree) ensureKeyType(k treeKey) error {
	if !t.keyTypeSet {
		t.keyType = k.Type
		t.keyTypeSet = true
		return nil
	}
	if t.keyType != k.Type {
		return &RuntimeError{Message: "tree key type mismatch"}
	}
	return nil
}

func (t *Tree) Set(key Value, val Value) error {
	k, err := treeKeyForValue(key)
	if err != nil {
		return err
	}
	if err := t.ensureKeyType(k); err != nil {
		return err
	}

	added := false
	switch t.kind {
	case "avl":
		t.root = avlInsert(t.root, k, val, &added)
	case "treap":
		t.root = treapInsert(t.root, k, val, &added)
	default:
		return &RuntimeError{Message: "unsupported tree kind"}
	}
	if added {
		t.size++
	}
	return nil
}

func (t *Tree) findNodeByValueKey(key Value) (*treeNode, error) {
	k, err := treeKeyForValue(key)
	if err != nil {
		return nil, err
	}
	if !t.keyTypeSet {
		return nil, nil
	}
	if t.keyType != k.Type {
		return nil, &RuntimeError{Message: "tree key type mismatch"}
	}
	return findTreeNode(t.root, k)
}

func findTreeNode(root *treeNode, key treeKey) (*treeNode, error) {
	cur := root
	for cur != nil {
		c, err := compareTreeKey(key, cur.key)
		if err != nil {
			return nil, err
		}
		if c < 0 {
			cur = cur.left
		} else if c > 0 {
			cur = cur.right
		} else {
			return cur, nil
		}
	}
	return nil, nil
}

func (t *Tree) Get(key Value) (Value, error) {
	node, err := t.findNodeByValueKey(key)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return NullValue, nil
	}
	return node.value, nil
}

func (t *Tree) Has(key Value) (bool, error) {
	node, err := t.findNodeByValueKey(key)
	if err != nil {
		return false, err
	}
	return node != nil, nil
}

func (t *Tree) Delete(key Value) (bool, error) {
	k, err := treeKeyForValue(key)
	if err != nil {
		return false, err
	}
	if !t.keyTypeSet {
		return false, nil
	}
	if t.keyType != k.Type {
		return false, &RuntimeError{Message: "tree key type mismatch"}
	}

	deleted := false
	switch t.kind {
	case "avl":
		t.root = avlDelete(t.root, k, &deleted)
	case "treap":
		t.root, deleted = treapDelete(t.root, k)
	default:
		return false, &RuntimeError{Message: "unsupported tree kind"}
	}
	if deleted {
		t.size--
		if t.size == 0 {
			t.keyTypeSet = false
			t.keyType = ""
		}
	}
	return deleted, nil
}

func (t *Tree) MinNode() *treeNode {
	cur := t.root
	for cur != nil && cur.left != nil {
		cur = cur.left
	}
	return cur
}

func (t *Tree) MaxNode() *treeNode {
	cur := t.root
	for cur != nil && cur.right != nil {
		cur = cur.right
	}
	return cur
}

func (t *Tree) LowerBound(key Value) (*treeNode, error) {
	k, err := treeKeyForValue(key)
	if err != nil {
		return nil, err
	}
	if !t.keyTypeSet {
		return nil, nil
	}
	if t.keyType != k.Type {
		return nil, &RuntimeError{Message: "tree key type mismatch"}
	}

	var best *treeNode
	cur := t.root
	for cur != nil {
		c, err := compareTreeKey(cur.key, k)
		if err != nil {
			return nil, err
		}
		if c >= 0 {
			best = cur
			cur = cur.left
		} else {
			cur = cur.right
		}
	}
	return best, nil
}

func (t *Tree) UpperBound(key Value) (*treeNode, error) {
	k, err := treeKeyForValue(key)
	if err != nil {
		return nil, err
	}
	if !t.keyTypeSet {
		return nil, nil
	}
	if t.keyType != k.Type {
		return nil, &RuntimeError{Message: "tree key type mismatch"}
	}

	var best *treeNode
	cur := t.root
	for cur != nil {
		c, err := compareTreeKey(cur.key, k)
		if err != nil {
			return nil, err
		}
		if c > 0 {
			best = cur
			cur = cur.left
		} else {
			cur = cur.right
		}
	}
	return best, nil
}

func (t *Tree) Floor(key Value) (*treeNode, error) {
	k, err := treeKeyForValue(key)
	if err != nil {
		return nil, err
	}
	if !t.keyTypeSet {
		return nil, nil
	}
	if t.keyType != k.Type {
		return nil, &RuntimeError{Message: "tree key type mismatch"}
	}

	var best *treeNode
	cur := t.root
	for cur != nil {
		c, err := compareTreeKey(cur.key, k)
		if err != nil {
			return nil, err
		}
		if c <= 0 {
			best = cur
			cur = cur.right
		} else {
			cur = cur.left
		}
	}
	return best, nil
}

func (t *Tree) Ceil(key Value) (*treeNode, error) {
	return t.LowerBound(key)
}

func (t *Tree) Predecessor(key Value) (*treeNode, error) {
	return t.FloorExclusive(key)
}

func (t *Tree) Successor(key Value) (*treeNode, error) {
	return t.UpperBound(key)
}

func (t *Tree) FloorExclusive(key Value) (*treeNode, error) {
	k, err := treeKeyForValue(key)
	if err != nil {
		return nil, err
	}
	if !t.keyTypeSet {
		return nil, nil
	}
	if t.keyType != k.Type {
		return nil, &RuntimeError{Message: "tree key type mismatch"}
	}

	var best *treeNode
	cur := t.root
	for cur != nil {
		c, err := compareTreeKey(cur.key, k)
		if err != nil {
			return nil, err
		}
		if c < 0 {
			best = cur
			cur = cur.right
		} else {
			cur = cur.left
		}
	}
	return best, nil
}

func (t *Tree) Closest(key Value, tieUpper bool) (*treeNode, bool, error) {
	k, err := treeKeyForValue(key)
	if err != nil {
		return nil, false, err
	}
	if !t.keyTypeSet {
		return nil, false, nil
	}
	if t.keyType != k.Type {
		return nil, false, &RuntimeError{Message: "tree key type mismatch"}
	}
	if t.keyType != INTEGER && t.keyType != FLOAT {
		return nil, false, &RuntimeError{Message: "closest expects numeric tree keys (int or float)"}
	}

	lower, err := t.Floor(key)
	if err != nil {
		return nil, false, err
	}
	upper, err := t.Ceil(key)
	if err != nil {
		return nil, false, err
	}

	if lower == nil && upper == nil {
		return nil, false, nil
	}
	if lower == nil {
		return upper, false, nil
	}
	if upper == nil {
		return lower, false, nil
	}
	if c, err := compareTreeKey(lower.key, upper.key); err == nil && c == 0 {
		return lower, true, nil
	}

	target := numericTreeKey(k)
	lowerDist := math.Abs(target - numericTreeKey(lower.key))
	upperDist := math.Abs(numericTreeKey(upper.key) - target)
	if upperDist < lowerDist {
		return upper, false, nil
	}
	if lowerDist < upperDist {
		return lower, false, nil
	}
	if tieUpper {
		return upper, false, nil
	}
	return lower, false, nil
}

func numericTreeKey(k treeKey) float64 {
	switch k.Type {
	case INTEGER:
		return float64(k.I)
	case FLOAT:
		return k.F
	default:
		return 0
	}
}

func (t *Tree) Range(from Value, to Value, includeFrom bool, includeTo bool, limit int) ([]Value, error) {
	fromKey, err := treeKeyForValue(from)
	if err != nil {
		return nil, err
	}
	toKey, err := treeKeyForValue(to)
	if err != nil {
		return nil, err
	}
	if !t.keyTypeSet {
		return []Value{}, nil
	}
	if t.keyType != fromKey.Type || t.keyType != toKey.Type {
		return nil, &RuntimeError{Message: "tree key type mismatch"}
	}
	order, err := compareTreeKey(fromKey, toKey)
	if err != nil {
		return nil, err
	}
	if order > 0 {
		return []Value{}, nil
	}

	items := make([]Value, 0)
	var visit func(*treeNode) error
	visit = func(n *treeNode) error {
		if n == nil {
			return nil
		}
		if limit > 0 && len(items) >= limit {
			return nil
		}

		cmpFrom, err := compareTreeKey(n.key, fromKey)
		if err != nil {
			return err
		}
		cmpTo, err := compareTreeKey(n.key, toKey)
		if err != nil {
			return err
		}

		if cmpFrom > 0 {
			if err := visit(n.left); err != nil {
				return err
			}
		}
		if limit > 0 && len(items) >= limit {
			return nil
		}

		geFrom := cmpFrom > 0 || (cmpFrom == 0 && includeFrom)
		leTo := cmpTo < 0 || (cmpTo == 0 && includeTo)
		if geFrom && leTo {
			items = append(items, treeItemValue(n))
			if limit > 0 && len(items) >= limit {
				return nil
			}
		}
		if limit > 0 && len(items) >= limit {
			return nil
		}

		if cmpTo < 0 {
			if err := visit(n.right); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(t.root); err != nil {
		return nil, err
	}
	return items, nil
}

func (t *Tree) Keys() []Value {
	out := []Value{}
	inOrderTree(t.root, func(n *treeNode) {
		out = append(out, treeKeyToValue(n.key))
	})
	return out
}

func (t *Tree) Values() []Value {
	out := []Value{}
	inOrderTree(t.root, func(n *treeNode) {
		out = append(out, n.value)
	})
	return out
}

func (t *Tree) Items() []Value {
	out := []Value{}
	inOrderTree(t.root, func(n *treeNode) {
		out = append(out, treeItemValue(n))
	})
	return out
}

func (t *Tree) MaxDepth() int {
	return treeDepth(t.root)
}

func treeDepth(n *treeNode) int {
	if n == nil {
		return 0
	}
	left := treeDepth(n.left)
	right := treeDepth(n.right)
	if left > right {
		return left + 1
	}
	return right + 1
}

func (t *Tree) MaxWidth() int {
	if t.root == nil {
		return 0
	}
	level := []*treeNode{t.root}
	max := 0
	for len(level) > 0 {
		if len(level) > max {
			max = len(level)
		}
		next := make([]*treeNode, 0, len(level)*2)
		for _, n := range level {
			if n.left != nil {
				next = append(next, n.left)
			}
			if n.right != nil {
				next = append(next, n.right)
			}
		}
		level = next
	}
	return max
}

func (t *Tree) Path(key Value) ([]Value, error) {
	k, err := treeKeyForValue(key)
	if err != nil {
		return nil, err
	}
	if !t.keyTypeSet {
		return nil, nil
	}
	if t.keyType != k.Type {
		return nil, &RuntimeError{Message: "tree key type mismatch"}
	}

	out := []Value{}
	cur := t.root
	for cur != nil {
		out = append(out, treeItemValue(cur))
		c, err := compareTreeKey(k, cur.key)
		if err != nil {
			return nil, err
		}
		if c == 0 {
			return out, nil
		}
		if c < 0 {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}
	return nil, nil
}

func (t *Tree) Rank(key Value) (int, bool, error) {
	k, err := treeKeyForValue(key)
	if err != nil {
		return 0, false, err
	}
	if !t.keyTypeSet {
		return 0, false, nil
	}
	if t.keyType != k.Type {
		return 0, false, &RuntimeError{Message: "tree key type mismatch"}
	}

	stack := []*treeNode{}
	cur := t.root
	count := 0
	for cur != nil || len(stack) > 0 {
		for cur != nil {
			stack = append(stack, cur)
			cur = cur.left
		}
		last := len(stack) - 1
		n := stack[last]
		stack = stack[:last]

		c, err := compareTreeKey(n.key, k)
		if err != nil {
			return 0, false, err
		}
		if c == 0 {
			return count, true, nil
		}
		if c > 0 {
			return 0, false, nil
		}
		count++
		cur = n.right
	}
	return 0, false, nil
}

func (t *Tree) Select(rank int) *treeNode {
	if rank < 0 || rank >= t.size {
		return nil
	}
	stack := []*treeNode{}
	cur := t.root
	idx := 0
	for cur != nil || len(stack) > 0 {
		for cur != nil {
			stack = append(stack, cur)
			cur = cur.left
		}
		last := len(stack) - 1
		n := stack[last]
		stack = stack[:last]
		if idx == rank {
			return n
		}
		idx++
		cur = n.right
	}
	return nil
}

type treeClosestCandidate struct {
	node     *treeNode
	exact    bool
	distance float64
	side     int // -1 lower, 0 exact, 1 upper
}

func (t *Tree) KClosest(key Value, k int, tieUpper bool) ([]treeClosestCandidate, error) {
	if k <= 0 {
		return []treeClosestCandidate{}, nil
	}
	targetKey, err := treeKeyForValue(key)
	if err != nil {
		return nil, err
	}
	if !t.keyTypeSet {
		return []treeClosestCandidate{}, nil
	}
	if t.keyType != targetKey.Type {
		return nil, &RuntimeError{Message: "tree key type mismatch"}
	}
	if t.keyType != INTEGER && t.keyType != FLOAT {
		return nil, &RuntimeError{Message: "kClosest expects numeric tree keys (int or float)"}
	}

	target := numericTreeKey(targetKey)
	candidates := []treeClosestCandidate{}
	inOrderTree(t.root, func(n *treeNode) {
		value := numericTreeKey(n.key)
		side := 0
		if value < target {
			side = -1
		} else if value > target {
			side = 1
		}
		candidates = append(candidates, treeClosestCandidate{
			node:     n,
			exact:    side == 0,
			distance: math.Abs(value - target),
			side:     side,
		})
	})

	sort.Slice(candidates, func(i, j int) bool {
		a := candidates[i]
		b := candidates[j]
		if a.distance < b.distance {
			return true
		}
		if a.distance > b.distance {
			return false
		}
		if a.side != b.side {
			if a.side == 0 {
				return true
			}
			if b.side == 0 {
				return false
			}
			if (a.side == -1 && b.side == 1) || (a.side == 1 && b.side == -1) {
				if tieUpper {
					return a.side == 1
				}
				return a.side == -1
			}
		}
		c, _ := compareTreeKey(a.node.key, b.node.key)
		return c < 0
	})

	if k > len(candidates) {
		k = len(candidates)
	}
	return candidates[:k], nil
}

func treeDistanceValue(distance float64, keyType ValueType) Value {
	if keyType == INTEGER {
		return &Integer{Value: int64(distance)}
	}
	return &Float{Value: distance}
}

func (t *Tree) PopMin() (Value, bool, error) {
	n := t.MinNode()
	if n == nil {
		return NullValue, false, nil
	}
	item := treeItemValue(n)
	_, err := t.Delete(treeKeyToValue(n.key))
	if err != nil {
		return nil, false, err
	}
	return item, true, nil
}

func (t *Tree) PopMax() (Value, bool, error) {
	n := t.MaxNode()
	if n == nil {
		return NullValue, false, nil
	}
	item := treeItemValue(n)
	_, err := t.Delete(treeKeyToValue(n.key))
	if err != nil {
		return nil, false, err
	}
	return item, true, nil
}

func (t *Tree) Validate() (bool, string) {
	if t == nil {
		return false, "tree is null"
	}
	if t.root == nil {
		if t.size != 0 {
			return false, "size mismatch: empty root with non-zero size"
		}
		if t.keyTypeSet {
			return false, "empty tree must not keep key type"
		}
		return true, ""
	}
	if t.kind != "avl" && t.kind != "treap" {
		return false, "unsupported tree kind"
	}
	if !t.keyTypeSet {
		return false, "non-empty tree must have key type"
	}
	count, reason := t.validateNode(t.root, nil, nil)
	if reason != "" {
		return false, reason
	}
	if count != t.size {
		return false, "size mismatch: node count differs from tree size"
	}
	return true, ""
}

func (t *Tree) validateNode(n *treeNode, min *treeKey, max *treeKey) (int, string) {
	if n == nil {
		return 0, ""
	}
	if n.key.Type != t.keyType {
		return 0, "key type mismatch inside tree"
	}
	if min != nil {
		c, err := compareTreeKey(n.key, *min)
		if err != nil {
			return 0, err.Error()
		}
		if c <= 0 {
			return 0, "invalid BST ordering (left bound)"
		}
	}
	if max != nil {
		c, err := compareTreeKey(n.key, *max)
		if err != nil {
			return 0, err.Error()
		}
		if c >= 0 {
			return 0, "invalid BST ordering (right bound)"
		}
	}

	leftCount, leftReason := t.validateNode(n.left, min, &n.key)
	if leftReason != "" {
		return 0, leftReason
	}
	rightCount, rightReason := t.validateNode(n.right, &n.key, max)
	if rightReason != "" {
		return 0, rightReason
	}

	if t.kind == "avl" {
		expectedHeight := nodeHeight(n.left)
		rightHeight := nodeHeight(n.right)
		if rightHeight > expectedHeight {
			expectedHeight = rightHeight
		}
		expectedHeight++
		if n.height != expectedHeight {
			return 0, "invalid AVL node height"
		}
		b := balanceFactor(n)
		if b < -1 || b > 1 {
			return 0, "invalid AVL balance factor"
		}
	}
	if t.kind == "treap" {
		if n.left != nil && n.left.priority < n.priority {
			return 0, "invalid treap heap property (left child priority)"
		}
		if n.right != nil && n.right.priority < n.priority {
			return 0, "invalid treap heap property (right child priority)"
		}
	}
	return 1 + leftCount + rightCount, ""
}

func (t *Tree) BulkLoad(items []Value, replace bool) error {
	if replace {
		t.root = nil
		t.size = 0
		t.keyType = ""
		t.keyTypeSet = false
	}
	for _, item := range items {
		key, value, err := treeBulkItemKV(item)
		if err != nil {
			return err
		}
		if err := t.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func treeBulkItemKV(item Value) (Value, Value, error) {
	switch x := item.(type) {
	case *Object:
		key, ok := x.Pairs["key"]
		if !ok {
			return nil, nil, &RuntimeError{Message: "bulkLoad object item requires key"}
		}
		value, ok := x.Pairs["value"]
		if !ok {
			return nil, nil, &RuntimeError{Message: "bulkLoad object item requires value"}
		}
		return key, value, nil
	case *Array:
		if len(x.Elements) != 2 {
			return nil, nil, &RuntimeError{Message: "bulkLoad tuple item must have 2 elements"}
		}
		return x.Elements[0], x.Elements[1], nil
	default:
		return nil, nil, &RuntimeError{Message: "bulkLoad items must be objects {key,value} or [key, value]"}
	}
}

func inOrderTree(root *treeNode, visit func(*treeNode)) {
	if root == nil {
		return
	}
	inOrderTree(root.left, visit)
	visit(root)
	inOrderTree(root.right, visit)
}

func treeItemValue(n *treeNode) Value {
	if n == nil {
		return NullValue
	}
	return &Object{Pairs: map[string]Value{
		"key":   treeKeyToValue(n.key),
		"value": n.value,
	}}
}

// AVL
func nodeHeight(n *treeNode) int {
	if n == nil {
		return 0
	}
	return n.height
}

func updateHeight(n *treeNode) {
	if n == nil {
		return
	}
	lh := nodeHeight(n.left)
	rh := nodeHeight(n.right)
	if lh > rh {
		n.height = lh + 1
	} else {
		n.height = rh + 1
	}
}

func balanceFactor(n *treeNode) int {
	if n == nil {
		return 0
	}
	return nodeHeight(n.left) - nodeHeight(n.right)
}

func rotateLeft(z *treeNode) *treeNode {
	y := z.right
	T2 := y.left
	y.left = z
	z.right = T2
	updateHeight(z)
	updateHeight(y)
	return y
}

func rotateRight(z *treeNode) *treeNode {
	y := z.left
	T3 := y.right
	y.right = z
	z.left = T3
	updateHeight(z)
	updateHeight(y)
	return y
}

func avlInsert(root *treeNode, key treeKey, val Value, added *bool) *treeNode {
	if root == nil {
		*added = true
		return &treeNode{key: key, value: val, height: 1}
	}

	c, _ := compareTreeKey(key, root.key)
	if c < 0 {
		root.left = avlInsert(root.left, key, val, added)
	} else if c > 0 {
		root.right = avlInsert(root.right, key, val, added)
	} else {
		root.value = val
		return root
	}

	updateHeight(root)
	b := balanceFactor(root)

	// LL
	if b > 1 {
		c2, _ := compareTreeKey(key, root.left.key)
		if c2 < 0 {
			return rotateRight(root)
		}
		// LR
		root.left = rotateLeft(root.left)
		return rotateRight(root)
	}
	// RR / RL
	if b < -1 {
		c2, _ := compareTreeKey(key, root.right.key)
		if c2 > 0 {
			return rotateLeft(root)
		}
		root.right = rotateRight(root.right)
		return rotateLeft(root)
	}
	return root
}

func minTreeNode(n *treeNode) *treeNode {
	cur := n
	for cur != nil && cur.left != nil {
		cur = cur.left
	}
	return cur
}

func avlDelete(root *treeNode, key treeKey, deleted *bool) *treeNode {
	if root == nil {
		return nil
	}
	c, _ := compareTreeKey(key, root.key)
	if c < 0 {
		root.left = avlDelete(root.left, key, deleted)
	} else if c > 0 {
		root.right = avlDelete(root.right, key, deleted)
	} else {
		*deleted = true
		if root.left == nil || root.right == nil {
			if root.left != nil {
				return root.left
			}
			return root.right
		}
		succ := minTreeNode(root.right)
		root.key = succ.key
		root.value = succ.value
		root.right = avlDelete(root.right, succ.key, deleted)
	}

	updateHeight(root)
	b := balanceFactor(root)
	if b > 1 {
		if balanceFactor(root.left) >= 0 {
			return rotateRight(root)
		}
		root.left = rotateLeft(root.left)
		return rotateRight(root)
	}
	if b < -1 {
		if balanceFactor(root.right) <= 0 {
			return rotateLeft(root)
		}
		root.right = rotateRight(root.right)
		return rotateLeft(root)
	}
	return root
}

// Treap
func treapInsert(root *treeNode, key treeKey, val Value, added *bool) *treeNode {
	if root == nil {
		*added = true
		return &treeNode{
			key:      key,
			value:    val,
			priority: rand.Int(),
		}
	}
	c, _ := compareTreeKey(key, root.key)
	if c < 0 {
		root.left = treapInsert(root.left, key, val, added)
		if root.left != nil && root.left.priority < root.priority {
			root = rotateRight(root)
		}
		return root
	}
	if c > 0 {
		root.right = treapInsert(root.right, key, val, added)
		if root.right != nil && root.right.priority < root.priority {
			root = rotateLeft(root)
		}
		return root
	}
	root.value = val
	return root
}

func treapDelete(root *treeNode, key treeKey) (*treeNode, bool) {
	if root == nil {
		return nil, false
	}
	c, _ := compareTreeKey(key, root.key)
	if c < 0 {
		var deleted bool
		root.left, deleted = treapDelete(root.left, key)
		return root, deleted
	}
	if c > 0 {
		var deleted bool
		root.right, deleted = treapDelete(root.right, key)
		return root, deleted
	}
	return treapMerge(root.left, root.right), true
}

func treapMerge(a, b *treeNode) *treeNode {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.priority < b.priority {
		a.right = treapMerge(a.right, b)
		return a
	}
	b.left = treapMerge(a, b.left)
	return b
}
