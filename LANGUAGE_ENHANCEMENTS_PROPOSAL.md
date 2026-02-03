# Karl Language Enhancement Proposals

**Date:** 2026-02-03  
**Context:** Workflow Engine Development - Import Normalization & Boolean Strictness Issues

## Executive Summary

During the development of the enhanced workflow engine, we encountered two significant language limitations that significantly impact code readability, maintainability, and developer experience:

1. **Strict Boolean Requirement in Conditionals** - All `if` conditions must be explicit boolean comparisons
2. **No Dynamic Map/Object Property Assignment** - Cannot use `map[key] = value` pattern for building maps

This document proposes language enhancements to address these limitations while maintaining Karl's type safety and predictability.

---

## Proposal 1: Truthy/Falsy Support for Conditionals

### Current Limitation

Karl currently requires explicit boolean comparisons in all conditional statements:

```karl
// ❌ Current - NOT ALLOWED
if config.retryPolicy {
    // use retry policy
}

if storage {
    // use storage
}

if !isCompleted {
    // process incomplete
}

// ✅ Current - REQUIRED
if config.retryPolicy != null {
    // use retry policy
}

if storage != null {
    // use storage
}

if isCompleted == false {
    // process incomplete
}
```

### Impact on Code Quality

This strictness leads to:
- **Verbose code**: Every conditional requires explicit comparison operators
- **Reduced readability**: Common patterns like null-checking become cluttered
- **Cognitive overhead**: Developers must remember to add `!= null` or `== true` everywhere
- **Maintenance burden**: 100+ lines of code changes needed for boolean fixes in workflow engine

### Proposed Enhancement: Truthy/Falsy Semantics

Introduce JavaScript/Python-style truthy/falsy evaluation with clear, predictable rules:

#### Truthy/Falsy Rules

**Falsy Values:**
- `null`
- `false`
- `0` (number zero)
- `""` (empty string)
- `[]` (empty array)
- `{}` (empty object/map)

**Truthy Values:**
- Everything else
- `true`
- Non-zero numbers
- Non-empty strings
- Non-empty arrays
- Non-empty objects/maps

#### Examples

```karl
// ✅ Proposed - Natural null checking
if config.retryPolicy {
    // Truthy if retryPolicy exists (not null)
    let retryEngine = RetryModule.createRetryEngine(config.retryPolicy)
}

// ✅ Proposed - Natural boolean checking
if result.success {
    // Truthy if success is true
    log("Success!")
}

// ✅ Proposed - Natural negation
if !isCompleted {
    // Truthy if isCompleted is false or null
    processIncomplete()
}

// ✅ Proposed - Natural existence checking
if storage {
    // Truthy if storage exists (not null)
    storage.save(state)
}

// ✅ Proposed - Empty checking
if tasks.length {
    // Truthy if array has elements
    processTasks(tasks)
}
```

### Backward Compatibility

**Option A: Opt-in via pragma**
```karl
// Enable truthy/falsy semantics for this file
#pragma truthy

if config.retryPolicy {  // Now allowed
    // ...
}
```

**Option B: New language version**
```karl
// Karl 2.0 - truthy/falsy enabled by default
// Karl 1.0 - strict boolean only (current behavior)
```

**Option C: Gradual migration**
- Emit warnings for implicit truthy/falsy usage in Karl 1.x
- Make it the default in Karl 2.0

### Benefits

1. **Improved Readability**: Code reads more naturally
2. **Reduced Verbosity**: ~30-40% fewer characters in conditional expressions
3. **Better DX**: Matches expectations from other modern languages
4. **Easier Onboarding**: More intuitive for developers from JS/Python/Ruby backgrounds

### Type Safety Considerations

To maintain Karl's type safety:

1. **Explicit Comparisons Still Allowed**: `if x == true` remains valid and preferred for clarity
2. **Linter Rules**: Configurable rules to enforce explicit comparisons where desired
3. **Type Inference**: Compiler can still infer types and catch errors
4. **Warning Mode**: Optional warnings for implicit truthy/falsy conversions

---

## Proposal 2: Dynamic Map/Object Property Assignment

### Current Limitation

Karl does not support dynamic property assignment on maps/objects:

```karl
// ❌ Current - NOT ALLOWED
let map = {}
map[key] = value  // Runtime error: "index assignment requires array"

// ❌ Current - NOT ALLOWED
for i < nodes.length with i = 0, deps = {} {
    deps[nodes[i].id] = []  // Runtime error
    i = i + 1
} then deps
```

### Impact on Code Quality

This limitation makes it **impossible** to:
- Build maps/dictionaries dynamically
- Create lookup tables from arrays
- Implement common data structure patterns (dependency graphs, state machines, caches)
- Write idiomatic code for many algorithms

### Current Workarounds (All Inadequate)

**Workaround 1: Pre-define all keys** (Not feasible for dynamic data)
```karl
let map = {
    knownKey1: value1,
    knownKey2: value2,
    // But what if keys are dynamic?
}
```

**Workaround 2: Use arrays with linear search** (O(n) instead of O(1))
```karl
let pairs = [
    { key: "a", value: 1 },
    { key: "b", value: 2 },
]
// Linear search required for lookups - very inefficient
```

**Workaround 3: Rebuild entire map each iteration** (Extremely inefficient)
```karl
for i < n with i = 0, map = {} {
    map = {
        ...map,  // Spread entire map
        [newKey]: newValue,  // Add one property
    }
} then map
// O(n²) complexity instead of O(n)
```

### Proposed Enhancement: Dynamic Property Assignment

Enable the `map[key] = value` pattern for objects/maps:

```karl
// ✅ Proposed - Dynamic property assignment
let map = {}
map["dynamicKey"] = "value"
map[variableKey] = computedValue

// ✅ Proposed - Building maps in loops
let dependencies = {}
for i < edges.length with i = 0, deps = {} {
    let edge = edges[i]
    deps[edge.target] = deps[edge.target] + [edge.source]
    i = i + 1
} then deps

// ✅ Proposed - State management
let state = {}
for i < nodes.length with i = 0, s = {} {
    s[nodes[i].id] = false  // Initialize state
    i = i + 1
} then s
```

### Implementation Approaches

**Option A: Mutable Maps (Simplest)**
```karl
let map = {}  // Creates mutable map
map[key] = value  // Mutates in place
```

**Option B: Immutable with Update Syntax**
```karl
let map1 = {}
let map2 = map1 with { [key]: value }  // Returns new map
```

**Option C: Explicit Mutable Type**
```karl
let map = mutableMap()  // Explicit mutable type
map[key] = value  // Allowed on mutable maps only

let immutableMap = {}  // Regular maps remain immutable
immutableMap[key] = value  // Error: immutable map
```

### Recommended Approach: Option A with Constraints

Allow mutable maps but with clear semantics:

1. **Loop-scoped Mutability**: Maps created in loop `with` clauses are mutable within that loop
2. **Immutable by Default**: Maps created outside loops remain immutable
3. **Explicit Conversion**: `toMutable(map)` and `toImmutable(map)` functions

```karl
// Immutable by default
let config = { a: 1 }
config[" b"] = 2  // Error: cannot mutate immutable map

// Mutable in loops
for i < n with i = 0, map = {} {
    map[key] = value  // OK: loop-scoped mutable map
} then map  // Returns immutable snapshot

// Explicit mutable
let cache = toMutable({})
cache[key] = value  // OK: explicitly mutable
```

### Benefits

1. **Enables Essential Patterns**: Dependency graphs, state machines, caches, lookup tables
2. **Performance**: O(1) map operations instead of O(n) array searches
3. **Idiomatic Code**: Matches patterns from other languages
4. **Practical Workflows**: Makes complex workflows (like DAG execution) actually possible

### Type Safety Considerations

1. **Type Inference**: Compiler infers map value types from assignments
2. **Homogeneous Maps**: All values must be same type (or use union types)
3. **Key Types**: Support string and number keys
4. **Runtime Checks**: Validate type consistency on assignment

---

## Proposal 3: Additional Quality-of-Life Improvements

### 3.1 Null-Coalescing Operator

```karl
// ✅ Proposed
let value = config.retryPolicy ?? defaultRetryPolicy
let id = config.workflowId ?? "workflow-" + str(now())

// Instead of current verbose ternary
let value = if config.retryPolicy != null { config.retryPolicy } else { defaultRetryPolicy }
```

### 3.2 Optional Chaining

```karl
// ✅ Proposed
let result = config.retryPolicy?.maxAttempts ?? 3

// Instead of current nested checks
let result = if config.retryPolicy != null {
    if config.retryPolicy.maxAttempts != null {
        config.retryPolicy.maxAttempts
    } else {
        3
    }
} else {
    3
}
```

### 3.3 Safe Property Access

```karl
// ✅ Proposed - Returns null if property doesn't exist
let value = config?.unknownProperty  // null instead of error

// Current - Throws error
let value = config.unknownProperty  // Runtime error: missing property
```

---

## Implementation Priority

### Phase 1: Critical (Blocks Workflow Engine)
1. **Dynamic Map Assignment** - Required for DAG execution
2. **Truthy/Falsy Support** - Significantly improves code quality

### Phase 2: High Value
3. **Null-Coalescing Operator** - Common pattern
4. **Safe Property Access** - Prevents runtime errors

### Phase 3: Nice to Have
5. **Optional Chaining** - Convenience feature

---

## Migration Strategy

### For Existing Code

1. **Backward Compatible**: All existing code continues to work
2. **Opt-in Features**: Use pragma or version flag to enable new features
3. **Gradual Adoption**: Teams can migrate file-by-file
4. **Linter Support**: Automated migration tools

### For New Code

1. **Recommended Defaults**: New projects use enhanced semantics
2. **Style Guide**: Document best practices
3. **Examples**: Update all examples to demonstrate new features

---

## Conclusion

These enhancements would:

1. **Unblock Critical Features**: Enable DAG execution and complex workflows
2. **Improve Developer Experience**: More readable, maintainable code
3. **Maintain Type Safety**: Preserve Karl's strengths while adding flexibility
4. **Align with Modern Languages**: Match developer expectations from JS/Python/Go

### Immediate Action Items

1. Review and approve proposals
2. Implement dynamic map assignment (critical blocker)
3. Implement truthy/falsy support (major DX improvement)
4. Update language specification
5. Update compiler/interpreter
6. Provide migration guide and examples

---

## Appendix: Real-World Impact

### Before (Current Karl)

```karl
// Building a dependency map - IMPOSSIBLE in current Karl
let dependencies = for i < edges.length with i = 0, deps = {} {
    let edge = edges[i]
    let currentDeps = deps[edge.target]  // Error: missing property
    deps[edge.target] = currentDeps + [edge.source]  // Error: index assignment requires array
    i = i + 1
} then deps
```

### After (With Proposals)

```karl
// Building a dependency map - NATURAL and EFFICIENT
let dependencies = for i < edges.length with i = 0, deps = {} {
    let edge = edges[i]
    let currentDeps = deps[edge.target] ?? []  // Null-coalescing
    deps[edge.target] = currentDeps + [edge.source]  // Dynamic assignment
    i = i + 1
} then deps

// Using the map - NATURAL and READABLE
if dependencies[nodeId] {  // Truthy check
    processDependencies(dependencies[nodeId])
}
```

### Code Reduction Example

From workflow engine refactoring:
- **Before**: 150+ explicit boolean comparisons (`== true`, `!= null`, `== false`)
- **After**: ~50 natural truthy/falsy checks
- **Savings**: ~100 lines of boilerplate, 40% reduction in conditional verbosity
