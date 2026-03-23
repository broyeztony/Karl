package interpreter

import (
	"math/rand"
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
