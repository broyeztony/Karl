# Karl Examples

This folder contains focused, single-feature examples plus a few larger programs.
Each example is intended to be standalone and to demonstrate one specific feature.

## Feature Examples

- `examples/features/if_else.k` - if/else expressions
- `examples/features/match_basic.k` - match with range and wildcard
- `examples/features/loop_for.k` - for loops with break/continue
- `examples/features/functions_basic.k` - basic functions
- `examples/features/recursion.k` - recursion
- `examples/features/closures.k` - closures
- `examples/features/lists_basic.k` - arrays + map/filter/reduce/sum/find/sort/length
- `examples/features/maps_basic.k` - map set/get/has/delete/keys/values
- `examples/features/sets_basic.k` - set add/has/delete/values/size
- `examples/features/strings_basic.k` - string helpers
- `examples/features/objects_basic.k` - object literals + spread
- `examples/features/object_disambiguation.k` - object vs block disambiguation
- `examples/features/struct_init.k` - struct init syntax sugar
- `examples/features/ranges_slices.k` - ranges and slices
- `examples/features/error_handling.k` - recoverable errors with `? {}` and `fail()`
- `examples/features/concurrency_basic.k` - `&`, `|`, `then`, `wait`
- `examples/features/concurrency_advanced.k` - rendezvous send/recv
- `examples/features/import_module.k` - module to import
- `examples/features/import_use.k` - import factory usage
- `examples/features/import_instances_module.k` - module with per-instance state
- `examples/features/import_instances.k` - multiple import instances
- `examples/features/query_basic.k` - query expressions
- `examples/features/equality.k` - `==` vs `eqv`

## Programs

- `examples/programs/loops_and_functions.k`
- `examples/programs/concurrency.k`
- `examples/programs/full_program.k`
- `examples/programs/pulse_pipeline.k`
- `examples/programs/game_tick.k`
- `examples/programs/retry_policy.k`

## Running

```
karl run examples/features/lists_basic.k
karl parse examples/features/lists_basic.k --format=json
```
