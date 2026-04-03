# Collections Examples

Focused collection examples for arrays, maps, sets, and ordered trees.

- `examples/features/collections/list_basic.k` - list map/filter/reduce/sort/find
- `examples/features/collections/map_basic.k` - map set/get/has/delete/keys/values
- `examples/features/collections/set_basic.k` - set uniqueness + membership
- `examples/features/collections/tree_basic.k` - ordered tree basics (AVL/Treap)
- `examples/features/collections/tree_pretty_print.k` - built-in ASCII tree rendering via `log(t)`
- `examples/features/collections/tree_closest_search.k` - nearest-key lookup via `closest`
- `examples/features/collections/tree_ordered_search.k` - floor/ceil/predecessor/successor/range
- `examples/features/collections/tree_order_stats.k` - `bulkLoad` + `rank/select/kClosest/popMin/popMax/validate`
- `examples/features/collections/tree_paths.k` - root-to-node paths (`tree.path` and `ntree.path`)
- `examples/features/collections/ntree_navigation.k` - hierarchical parent/children/siblings navigation
- `examples/features/collections/ntree_topology_metrics.k` - `depth/height/lca/pathBetween/subtreeSize`

Run:

```bash
karl run examples/features/collections/tree_basic.k
karl run examples/features/collections/tree_pretty_print.k
karl run examples/features/collections/tree_closest_search.k
karl run examples/features/collections/tree_ordered_search.k
karl run examples/features/collections/tree_order_stats.k
karl run examples/features/collections/tree_paths.k
karl run examples/features/collections/ntree_navigation.k
karl run examples/features/collections/ntree_topology_metrics.k
```
