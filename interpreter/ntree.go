package interpreter

import (
	"strings"
)

type nTreeNode struct {
	id        string
	parent    string
	hasParent bool
	value     Value
	children  []string
}

type NTree struct {
	rootID string
	nodes  map[string]*nTreeNode
}

func newNTree(rootID string, rootValue Value) (*NTree, error) {
	id := strings.TrimSpace(rootID)
	if id == "" {
		return nil, &RuntimeError{Message: "ntree root id must be non-empty string"}
	}
	if rootValue == nil {
		rootValue = NullValue
	}
	node := &nTreeNode{
		id:       id,
		value:    rootValue,
		children: []string{},
	}
	return &NTree{
		rootID: id,
		nodes: map[string]*nTreeNode{
			id: node,
		},
	}, nil
}

func (t *NTree) Type() ValueType { return NTREE }

func (t *NTree) Inspect() string {
	size := int64(0)
	if t != nil {
		size = int64(len(t.nodes))
	}
	if t == nil || t.rootID == "" || len(t.nodes) == 0 {
		return "ntree(size=" + (&Integer{Value: size}).Inspect() + ")"
	}
	head := "ntree(size=" + (&Integer{Value: size}).Inspect() + ", root=" + (&String{Value: t.rootID}).Inspect() + ")"
	root := t.nodes[t.rootID]
	if root == nil {
		return head
	}
	lines := []string{head}
	t.appendInspectLines(&lines, root, "", true)
	return strings.Join(lines, "\n")
}

func (t *NTree) appendInspectLines(lines *[]string, n *nTreeNode, prefix string, isLast bool) {
	if n == nil {
		return
	}
	branch := "|-- "
	nextPrefix := prefix + "|   "
	if isLast {
		branch = "`-- "
		nextPrefix = prefix + "    "
	}
	line := prefix + branch + (&String{Value: n.id}).Inspect() + ": " + n.value.Inspect()
	*lines = append(*lines, line)
	for i, childID := range n.children {
		child := t.nodes[childID]
		if child == nil {
			continue
		}
		t.appendInspectLines(lines, child, nextPrefix, i == len(n.children)-1)
	}
}

func (t *NTree) nodeByID(id string) (*nTreeNode, bool) {
	if t == nil {
		return nil, false
	}
	n, ok := t.nodes[id]
	return n, ok
}

func (t *NTree) nodeValueByID(id string) Value {
	n, ok := t.nodeByID(id)
	if !ok {
		return NullValue
	}
	return n.nodeValue()
}

func (n *nTreeNode) nodeValue() Value {
	parent := Value(NullValue)
	if n.hasParent {
		parent = &String{Value: n.parent}
	}
	children := make([]Value, 0, len(n.children))
	for _, childID := range n.children {
		children = append(children, &String{Value: childID})
	}
	return &Object{Pairs: map[string]Value{
		"id":       &String{Value: n.id},
		"parent":   parent,
		"value":    n.value,
		"children": &Array{Elements: children},
	}}
}

func (t *NTree) SetValue(id string, value Value) bool {
	n, ok := t.nodeByID(id)
	if !ok {
		return false
	}
	n.value = value
	return true
}

func (t *NTree) insertNode(parentID string, index int, childID string, value Value) error {
	parent, ok := t.nodeByID(parentID)
	if !ok {
		return &RuntimeError{Message: "unknown parent id: " + parentID}
	}
	childID = strings.TrimSpace(childID)
	if childID == "" {
		return &RuntimeError{Message: "child id must be non-empty string"}
	}
	if _, exists := t.nodes[childID]; exists {
		return &RuntimeError{Message: "node id already exists: " + childID}
	}
	if value == nil {
		value = NullValue
	}
	if index < 0 {
		index = len(parent.children)
	}
	if index > len(parent.children) {
		return &RuntimeError{Message: "insert index out of range"}
	}

	child := &nTreeNode{
		id:        childID,
		parent:    parentID,
		hasParent: true,
		value:     value,
		children:  []string{},
	}
	t.nodes[childID] = child
	parent.children = insertIDAt(parent.children, index, childID)
	return nil
}

func (t *NTree) Append(parentID string, childID string, value Value) error {
	return t.insertNode(parentID, -1, childID, value)
}

func (t *NTree) Prepend(parentID string, childID string, value Value) error {
	return t.insertNode(parentID, 0, childID, value)
}

func (t *NTree) InsertAt(parentID string, index int, childID string, value Value) error {
	return t.insertNode(parentID, index, childID, value)
}

func (t *NTree) Parent(id string) (*nTreeNode, bool) {
	n, ok := t.nodeByID(id)
	if !ok || !n.hasParent {
		return nil, false
	}
	parent, ok := t.nodeByID(n.parent)
	return parent, ok
}

func (t *NTree) Children(id string) ([]*nTreeNode, error) {
	n, ok := t.nodeByID(id)
	if !ok {
		return nil, &RuntimeError{Message: "unknown node id: " + id}
	}
	out := make([]*nTreeNode, 0, len(n.children))
	for _, childID := range n.children {
		child, ok := t.nodeByID(childID)
		if ok {
			out = append(out, child)
		}
	}
	return out, nil
}

func (t *NTree) Siblings(id string, includeSelf bool) ([]*nTreeNode, error) {
	n, ok := t.nodeByID(id)
	if !ok {
		return nil, &RuntimeError{Message: "unknown node id: " + id}
	}
	if !n.hasParent {
		return []*nTreeNode{}, nil
	}
	parent, ok := t.nodeByID(n.parent)
	if !ok {
		return []*nTreeNode{}, nil
	}
	out := make([]*nTreeNode, 0, len(parent.children))
	for _, childID := range parent.children {
		if !includeSelf && childID == id {
			continue
		}
		child, ok := t.nodeByID(childID)
		if ok {
			out = append(out, child)
		}
	}
	return out, nil
}

func (t *NTree) Ancestors(id string) ([]*nTreeNode, error) {
	n, ok := t.nodeByID(id)
	if !ok {
		return nil, &RuntimeError{Message: "unknown node id: " + id}
	}
	out := []*nTreeNode{}
	cur := n
	for cur.hasParent {
		p, ok := t.nodeByID(cur.parent)
		if !ok {
			break
		}
		out = append(out, p)
		cur = p
	}
	return out, nil
}

func (t *NTree) Path(id string) ([]*nTreeNode, bool) {
	n, ok := t.nodeByID(id)
	if !ok {
		return nil, false
	}
	path := []*nTreeNode{}
	cur := n
	path = append(path, cur)
	for cur.hasParent {
		p, ok := t.nodeByID(cur.parent)
		if !ok {
			return nil, false
		}
		path = append(path, p)
		cur = p
	}
	reverseNodes(path)
	return path, true
}

func (t *NTree) Depth(id string) (int, bool) {
	n, ok := t.nodeByID(id)
	if !ok {
		return 0, false
	}
	depth := 0
	cur := n
	for cur.hasParent {
		next, ok := t.nodeByID(cur.parent)
		if !ok {
			return 0, false
		}
		depth++
		cur = next
	}
	return depth, true
}

func (t *NTree) Height(id string) (int, bool) {
	n, ok := t.nodeByID(id)
	if !ok {
		return 0, false
	}
	return t.nodeHeight(n), true
}

func (t *NTree) nodeHeight(n *nTreeNode) int {
	if n == nil {
		return 0
	}
	if len(n.children) == 0 {
		return 0
	}
	maxHeight := 0
	for _, childID := range n.children {
		child, ok := t.nodeByID(childID)
		if !ok {
			continue
		}
		h := t.nodeHeight(child) + 1
		if h > maxHeight {
			maxHeight = h
		}
	}
	return maxHeight
}

func (t *NTree) LCA(a string, b string) (*nTreeNode, bool) {
	if _, ok := t.nodeByID(a); !ok {
		return nil, false
	}
	if _, ok := t.nodeByID(b); !ok {
		return nil, false
	}

	ancestors := map[string]struct{}{}
	cur, _ := t.nodeByID(a)
	for cur != nil {
		ancestors[cur.id] = struct{}{}
		if !cur.hasParent {
			break
		}
		next, ok := t.nodeByID(cur.parent)
		if !ok {
			return nil, false
		}
		cur = next
	}

	cur, _ = t.nodeByID(b)
	for cur != nil {
		if _, ok := ancestors[cur.id]; ok {
			return cur, true
		}
		if !cur.hasParent {
			break
		}
		next, ok := t.nodeByID(cur.parent)
		if !ok {
			return nil, false
		}
		cur = next
	}
	return nil, false
}

func (t *NTree) PathBetween(a string, b string) ([]*nTreeNode, bool) {
	pathA, ok := t.Path(a)
	if !ok {
		return nil, false
	}
	pathB, ok := t.Path(b)
	if !ok {
		return nil, false
	}
	if len(pathA) == 0 || len(pathB) == 0 {
		return nil, false
	}

	lcaIndex := -1
	limit := len(pathA)
	if len(pathB) < limit {
		limit = len(pathB)
	}
	for i := 0; i < limit; i++ {
		if pathA[i].id != pathB[i].id {
			break
		}
		lcaIndex = i
	}
	if lcaIndex < 0 {
		return nil, false
	}

	out := []*nTreeNode{}
	for i := len(pathA) - 1; i >= lcaIndex; i-- {
		out = append(out, pathA[i])
	}
	for i := lcaIndex + 1; i < len(pathB); i++ {
		out = append(out, pathB[i])
	}
	return out, true
}

func (t *NTree) SubtreeSize(id string) (int, bool) {
	root, ok := t.nodeByID(id)
	if !ok {
		return 0, false
	}
	count := 0
	stack := []*nTreeNode{root}
	for len(stack) > 0 {
		last := len(stack) - 1
		n := stack[last]
		stack = stack[:last]
		count++
		for i := len(n.children) - 1; i >= 0; i-- {
			child, ok := t.nodeByID(n.children[i])
			if ok {
				stack = append(stack, child)
			}
		}
	}
	return count, true
}

func (t *NTree) Descendants(id string, traversal string) ([]*nTreeNode, error) {
	root, ok := t.nodeByID(id)
	if !ok {
		return nil, &RuntimeError{Message: "unknown node id: " + id}
	}
	mode := strings.ToLower(strings.TrimSpace(traversal))
	if mode == "" {
		mode = "dfs"
	}
	if mode != "dfs" && mode != "bfs" {
		return nil, &RuntimeError{Message: "descendants traversal must be \"dfs\" or \"bfs\""}
	}
	out := []*nTreeNode{}
	if mode == "dfs" {
		stack := make([]string, 0, len(root.children))
		for i := len(root.children) - 1; i >= 0; i-- {
			stack = append(stack, root.children[i])
		}
		for len(stack) > 0 {
			last := len(stack) - 1
			nodeID := stack[last]
			stack = stack[:last]
			n, ok := t.nodeByID(nodeID)
			if !ok {
				continue
			}
			out = append(out, n)
			for i := len(n.children) - 1; i >= 0; i-- {
				stack = append(stack, n.children[i])
			}
		}
		return out, nil
	}

	queue := append([]string{}, root.children...)
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		n, ok := t.nodeByID(nodeID)
		if !ok {
			continue
		}
		out = append(out, n)
		queue = append(queue, n.children...)
	}
	return out, nil
}

func (t *NTree) Remove(id string, subtree bool) (bool, error) {
	n, ok := t.nodeByID(id)
	if !ok {
		return false, nil
	}
	if !subtree && !n.hasParent {
		return false, &RuntimeError{Message: "cannot remove root without subtree=true"}
	}

	if subtree {
		if !n.hasParent {
			t.nodes = map[string]*nTreeNode{}
			t.rootID = ""
			return true, nil
		}
		parent, ok := t.nodeByID(n.parent)
		if ok {
			parent.children = removeID(parent.children, id)
		}
		stack := []string{id}
		for len(stack) > 0 {
			last := len(stack) - 1
			nodeID := stack[last]
			stack = stack[:last]
			cur, ok := t.nodeByID(nodeID)
			if !ok {
				continue
			}
			stack = append(stack, cur.children...)
			delete(t.nodes, nodeID)
		}
		return true, nil
	}

	parent, ok := t.nodeByID(n.parent)
	if !ok {
		return false, &RuntimeError{Message: "invalid tree state: missing parent"}
	}
	idx := indexOfID(parent.children, id)
	if idx < 0 {
		return false, &RuntimeError{Message: "invalid tree state: missing child reference"}
	}
	parent.children = removeID(parent.children, id)
	for _, childID := range n.children {
		child, ok := t.nodeByID(childID)
		if !ok {
			continue
		}
		child.parent = parent.id
		child.hasParent = true
	}
	parent.children = insertIDsAt(parent.children, idx, n.children)
	delete(t.nodes, id)
	return true, nil
}

func (t *NTree) Move(id string, newParentID string, index int, hasIndex bool) error {
	n, ok := t.nodeByID(id)
	if !ok {
		return &RuntimeError{Message: "unknown node id: " + id}
	}
	if !n.hasParent {
		return &RuntimeError{Message: "cannot move root node"}
	}
	newParent, ok := t.nodeByID(newParentID)
	if !ok {
		return &RuntimeError{Message: "unknown parent id: " + newParentID}
	}
	if id == newParentID {
		return &RuntimeError{Message: "cannot move node under itself"}
	}
	if t.isAncestor(id, newParentID) {
		return &RuntimeError{Message: "cannot move node under its descendant"}
	}

	oldParent, ok := t.nodeByID(n.parent)
	if !ok {
		return &RuntimeError{Message: "invalid tree state: missing parent"}
	}
	oldIndex := indexOfID(oldParent.children, id)
	if oldIndex < 0 {
		return &RuntimeError{Message: "invalid tree state: missing child reference"}
	}
	oldParent.children = removeID(oldParent.children, id)

	target := len(newParent.children)
	if hasIndex {
		if index < 0 || index > len(newParent.children) {
			return &RuntimeError{Message: "move index out of range"}
		}
		target = index
		if oldParent.id == newParent.id && oldIndex < index {
			target = index - 1
		}
	}
	newParent.children = insertIDAt(newParent.children, target, id)
	n.parent = newParent.id
	n.hasParent = true
	return nil
}

func (t *NTree) isAncestor(ancestorID string, id string) bool {
	cur, ok := t.nodeByID(id)
	if !ok {
		return false
	}
	for cur.hasParent {
		if cur.parent == ancestorID {
			return true
		}
		next, ok := t.nodeByID(cur.parent)
		if !ok {
			return false
		}
		cur = next
	}
	return false
}

func insertIDAt(ids []string, index int, id string) []string {
	if index < 0 || index > len(ids) {
		index = len(ids)
	}
	ids = append(ids, "")
	copy(ids[index+1:], ids[index:])
	ids[index] = id
	return ids
}

func insertIDsAt(ids []string, index int, add []string) []string {
	if len(add) == 0 {
		return ids
	}
	if index < 0 || index > len(ids) {
		index = len(ids)
	}
	out := make([]string, 0, len(ids)+len(add))
	out = append(out, ids[:index]...)
	out = append(out, add...)
	out = append(out, ids[index:]...)
	return out
}

func removeID(ids []string, id string) []string {
	idx := indexOfID(ids, id)
	if idx < 0 {
		return ids
	}
	return append(ids[:idx], ids[idx+1:]...)
}

func indexOfID(ids []string, id string) int {
	for i, x := range ids {
		if x == id {
			return i
		}
	}
	return -1
}

func reverseNodes(nodes []*nTreeNode) {
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
}
