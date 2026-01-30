# Shapes (.shape files)

Shapes define *record-like* object structures for external data (JSON, HTTP, files, etc.).
They keep Karl objects as fixed-field records while supporting alias mapping for external keys
(e.g., `X-Amzn-Trace-Id` -> `xAmznTraceId`).

## Goals

- Preserve objects as **identifier-keyed records**.
- Allow **alias mapping** for external keys without changing object semantics.
- Validate/shape external data with **recoverable errors**.
- Support **round-trip encode/decode** with aliases in both directions.

## File format

- File extension: `.shape`
- Lines beginning with `//` are comments.
- Indentation defines nesting (4 spaces per level; tabs count as 4 spaces).

### Top-level declaration

```
ShapeName : Type
```

Multiple shapes can appear in the same file (one per top-level line).

Examples:
```
HttpResponse : object
Vehicules    : string[]
Color        : object
Colors       : Color[]
```

### Field declarations

```
[+|-] fieldName [ (alias) ] : Type
```

- `+` required field
- `-` optional field
- If omitted, fields are treated as **required** (same as using `+`).
- `fieldName` must be a valid Karl identifier.
- `alias` is optional text inside parentheses and may contain spaces or punctuation.
  It maps **external keys** to the internal `fieldName`.

Example:
```
+ xAmznTraceId(X-Amzn-Trace-Id) : string
- userAgent(User Agent)         : string
```

### Types

Supported primitives:
- `string`, `int`, `float`, `bool`, `null`, `any`

Containers:
- `object`
- `T[]` (array of T)

Named references:
- `TypeName` (reference another shape declared in the same file)
- References are resolved **within the same file** only.

If the declared type is `object`, nested field declarations must be indented beneath it.
If the declared type is `object[]`, nested field declarations describe the **array element shape**.

Example:
```
Colors : object[]
    + color : string
    + value : string
```

## Importing shapes

Shapes are imported directly (no factory call):

```
let HttpResponse = import "./shapes/HttpResponse.shape"
```

If the file declares **one** shape, the imported value is a **shape object** that can be used with `as`.
If the file declares **multiple** shapes, the imported value is an **object** whose properties
are shape values keyed by their declaration names.

Examples:
```
let Colors = import "./shapes/Colors.shape"
let shapes = import "./shapes/palette.shape"
let Color = shapes.Color
let parsed = decodeJson(res.body) as shapes.HttpResponse
```

Shape values can be used with `as`,
or called like a function to shape a value. `value as Shape` is
syntactic sugar for `Shape(value)`:

```
let parsed = HttpResponse(decodeJson(res.body)) ? { slideshow: { title: "fallback" } }
```

Both forms are equivalent.

Callable shape contract:
- A shape value accepts exactly **one argument** and returns the shaped value.
- Invalid inputs produce a **recoverable** error (usable with `? { ... }`).

## `as` operator (shaping)

`as` applies a shape to a value and returns a shaped value:

```
let parsed = decodeJson(res.body) as HttpResponse
```

- `as` is **recoverable**: validation errors produce a recoverable runtime error
  (so `? { ... }` can handle it).
- `as` performs **alias mapping** and **validation** based on the shape.
- `as` can be used with `?`:

```
let parsed = decodeJson(res.body) as HttpResponse ? { headers: { acceptEncoding: "gzip" }, }
```

### Precedence

`as` binds **tighter than** `?` so the shape is applied before recovery.

## Shaping semantics

Given `value as Shape`:

1) **Type check**
   - If `Shape` is `string`, `int`, `float`, `bool`, `null`:
     the value must match that type.
   - If `Shape` is `any`: no type check.
   - If `Shape` is `T[]`: the value must be an array and each element is shaped as `T`.
   - If `Shape` is `object`: the value must be an object or map-like value.

2) **Alias mapping (external -> internal)**
   - If a field has an alias, the external key is mapped to the internal identifier.
   - If both the internal key and alias key are present, **internal wins**.
   - Aliases are **only** for shaping external data; the shaped object uses internal field names.

3) **Required vs optional**
   - Missing **required** fields produce a recoverable error.
   - Missing **optional** fields are set to `null`.

4) **Extra fields**
   - Fields not declared in the shape are dropped.
     (Shape validation only enforces declared fields.)

## Encoding (internal -> external)

When a value carries shape metadata (from `as`), `encodeJson` should:

- use **alias names** for fields that declare an alias,
- otherwise use the internal identifier name.

This enables round-trip JSON where external keys like `User Agent`
map to internal `userAgent` and back.

## Example

```text
HttpResponse : object
    + headers : object
        + acceptEncoding(Accept-Encoding) : string
        + userAgent(User Agent)           : string
    + status  : int
```

```karl
let HttpResponse = import "./HttpResponse.shape"
let res = http({ method: "GET", url: "https://httpbin.org/get", headers: map(), })
let parsed = decodeJson(res.body) as HttpResponse
let ua = parsed.headers.userAgent
let out = encodeJson(parsed)
```
