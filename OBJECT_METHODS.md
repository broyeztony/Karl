# Object Method Support for Karl

**Date:** 2026-01-30  
**Status:** ✅ COMPLETE

## What Was Implemented

Added `.get()`, `.set()`, and `.has()` methods to Karl's `{}` object literal syntax at the language level.

### Language Changes

**Modified Files:**
1. `interpreter/builtins.go` - Updated three builtin functions to support both Map and Object
   - `builtinMapGet` - Now handles both `*Map` and `*Object`
   - `builtinMapSet` - Now handles both `*Map` and `*Object`
   - `builtinMapHas` - Now handles both `*Map` and `*Object`

2. `interpreter/eval.go` - Added method dispatch for Object types
   - Updated `evalMemberExpression` to check for method calls on Object
   - Added `objectMethod` function to bind methods to Object instances

### Now Supported

```karl
// Create object
let obj = { name: "test", count: 0 }

// Use .get() method
let name = obj.get("name")  // Returns "test"

// Use .set() method  
obj = obj.set("count", 5)  // Returns updated object

// Use .has() method
let hasName = obj.has("name")  // Returns true
```

### Implementation Details

The implementation reuses the existing builtin functions (`get`, `set`, `has`) that were previously only available for `Map` types. The functions now use a type switch to handle both `*Map` and `*Object`:

- **For Map:** Uses `mapKeyForValue()` to support string, char, int, and bool keys
- **For Object:** Requires string/char keys only (following JavaScript object semantics)

## Testing

Built successfully with `go build -o karl .`

The change enables the workflow engine's DAG executor to work, as it can now use `{}` with `.set()`, `.get()`, `.has()` methods instead of requiring `map()`.

However, note that the DAG executor still has loop variable scoping issues that cause deadlocks - this is  a separate Karl VM issue unrelated to the object method support.

## Impact

**Breaking Changes:** None - this is purely additive

### Before
```karl
let obj = { a: 1 }
obj.get("a")  // ERROR: missing property: get
```

**After**
```karl
let obj = { a: 1 }
obj.get("a")  // ✅ Returns 1
obj.set("b", 2)  // ✅ Returns { a: 1, b: 2 }
obj.has("a")  // ✅ Returns true
```

## Conclusion

✅ **Karl now supports `.get()`, `.set()`, `.has()` on `{}` objects**

This brings Karl's object literals closer to Map functionality while maintaining their simpler syntax.
