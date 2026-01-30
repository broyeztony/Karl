# Karl Shape Tests

This directory contains integration tests for the Karl Shape feature, written in Karl itself.

## Test Files

| Test File | Description | Focus Areas |
|-----------|-------------|-------------|
| `01_alias_collision.k` | Alias collision handling | Internal vs external key priority |
| `02_complex_aliases.k` | Complex alias names | Spaces, hyphens, special characters |
| `03_deep_nesting.k` | Deep object nesting | Required/optional at various levels |
| `04_array_validation.k` | Array validation | Empty, valid, invalid elements |
| `05_type_mismatches.k` | Type checking | All primitive type mismatches |
| `06_optional_fields.k` | Optional field behavior | Null assignment, mixed fields |
| `07_extra_fields_dropped.k` | Field filtering | Extra fields not in shape |

## Shape Fixtures

All test shapes are located in `fixtures/` directory:
- `Response.shape` - Simple alias testing
- `Headers.shape` - Complex HTTP header aliases
- `DeepStructure.shape` - 4-level deep nesting
- `Items.shape` - Array of objects
- `Numbers.shape` - Primitive array
- Plus various type-specific shapes

## Running Tests

### Run All Tests
```bash
./run_shape_tests.sh
```

### Run Individual Test
```bash
./karl run tests/shapes/01_alias_collision.k
./karl run tests/shapes/02_complex_aliases.k
```

### Run Specific Test Group
```bash
# Run only array tests
./karl run tests/shapes/04_array_validation.k

# Run only type tests
./karl run tests/shapes/05_type_mismatches.k
```

## Test Output

Tests use emoji indicators:
- ✅ `PASS` - Test assertion passed
- ❌ `FAIL` - Test assertion failed

Each test file logs its results and ends with a summary:
```
✅ PASS: Alias collision - internal key wins
✅ PASS: Alias mapping works with external key only
✅ PASS: Internal key works without alias
Test suite: Alias collision - Complete
```

## Edge Cases Covered

### 1. Alias Handling
- **Collision**: Internal key wins when both present
- **Complex names**: Spaces, hyphens, special characters
- **Case sensitivity**: Preserves exact alias names

### 2. Deep Nesting
- **Required fields**: Error if missing at any level
- **Optional fields**: Set to null if missing
- **Error propagation**: Deep errors trigger recovery

### 3. Array Validation
- **Empty arrays**: Valid for any array type
- **Element validation**: Each element checked
- **Partial failure**: One bad element fails entire array
- **Type checking**: Primitive arrays enforce element types

### 4. Type System
- **Strict checking**: String ≠ int ≠ float ≠ bool
- **Recoverable errors**: All type mismatches recoverable
- **Error operator**: Works with `? { fallback }`

### 5. Optional Fields
- **Null assignment**: Missing optional → null
- **Required validation**: Missing required → error
- **Mixed shapes**: Independent handling

### 6. Extra Fields
- **Top-level**: Dropped if not in shape
- **Nested**: Dropped at all levels
- **Preservation**: Only declared fields kept

## Test Philosophy

These tests are **integration tests** that:
1. Run through the actual Karl interpreter
2. Test the complete shape workflow (parse → import → apply → validate)
3. Verify end-user observable behavior
4. Document expected behavior with real examples

Complementary to the Go unit tests in:
- `interpreter/shape_edge_cases_test.go`
- `shape/parser_test.go`

## Adding New Tests

1. Create new test file: `tests/shapes/XX_test_name.k`
2. Create fixture shapes in: `tests/shapes/fixtures/YourShape.shape`
3. Follow the pattern:
   ```karl
   let Shape = import "tests/shapes/fixtures/YourShape.shape"
   
   let raw = "{ ... }"
   let parsed = decodeJson(raw) as Shape ? { fallback }
   
   if parsed.field == expected {
       log("✅ PASS: Description")
   } else {
       log("❌ FAIL: Description")
   }
   ```
4. Run tests with `./run_shape_tests.sh`

## Coverage

- **Total Karl Tests**: 7
- **Total Assertions**: ~35+
- **Shape Fixtures**: 14
- **Edge Cases**: All major scenarios from Go tests
- **Status**: ✅ All passing
