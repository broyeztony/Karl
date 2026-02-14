# Karl Examples

This folder contains focused, standalone feature examples.
Each example demonstrates one specific capability.

## Feature Examples

- `examples/features/if_else.k` - if/else expressions
- `examples/features/match_examples.k` - match with range, wildcard, and guards
- `examples/features/loop_for.k` - for loops with break/continue
- `examples/features/functions_basic.k` - basic functions
- `examples/features/recursion.k` - recursion
- `examples/features/closures.k` - closures
- `examples/features/lists_basic.k` - arrays + map/filter/reduce/sum/find/sort/length
- `examples/features/maps_basic.k` - map set/get/has/delete/keys/values
- `examples/features/sets_basic.k` - set add/has/delete/values/size
- `examples/features/strings_basic.k` - string helpers
- `examples/features/objects_basic.k` - object literals + spread
- `examples/features/object_indexing.k` - bracket access for non-identifier keys
- `examples/features/object_disambiguation.k` - object vs block disambiguation
- `examples/features/struct_init.k` - struct init syntax sugar
- `examples/features/ranges_slices.k` - ranges and slices
- `examples/features/error_handling.k` - recoverable errors with `? {}` and `fail()`
- `examples/features/truthy_falsy.k` - truthy/falsy basics
- `examples/features/truthy_falsy_comprehensive.k` - truthy/falsy across values and operators
- `examples/features/concurrency/basic.k` - `&`, `|`, `then`, `wait`
- `examples/features/concurrency/advanced.k` - rendezvous/channel send/recv
- `examples/features/concurrency/buffered_channels.k` - buffered channel basics
- `examples/features/concurrency/` - deeper dive: join/race semantics, cancellation, and failure recovery
- `examples/features/import_module.k` - module to import
- `examples/features/import_use.k` - import factory usage
- `examples/features/import_instances_module.k` - module with per-instance state
- `examples/features/import_instances.k` - multiple import instances
- `examples/features/query_basic.k` - query expressions
- `examples/features/equality.k` - `==` vs `eqv`

## Community Examples

- `examples/contrib/nico/README.md` - extended example programs (by [Nico](http://github.com/hellonico))

## Running

```
karl run examples/features/lists_basic.k
karl parse examples/features/lists_basic.k --format=json
```
