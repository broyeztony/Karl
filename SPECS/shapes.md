# Shapes (.shape files)

Shapes define *record-like* object structures for external data (JSON, HTTP, files, etc.).
They keep Karl objects as fixed-field records while supporting alias mapping for external keys
(e.g., `X-Amzn-Trace-Id` -> `xAmznTraceId`).

> Draft note: this spec includes a **codec-mapping** sketch (Option 4) for
> multi-format translation. It is for reasoning only and not implemented yet.

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

// ============================================
// Codec Mappings (Draft / Reasoning)
// ============================================

Single-field aliases are not enough when formats disagree on valid key names.
To keep one shape usable across **multiple codecs**, we propose **codec mapping blocks**
that live alongside shape definitions.

### Syntax (draft)

```
codec <format> <ShapeName>
    path = "externalKey"
```

- `format` is a codec identifier: `json`, `yaml`, `csv`, `tsv`, `query`, `headers`, ...
- `ShapeName` is a top-level shape in the same file.
- `path` is a **dot-separated field path** (nested objects).
- `externalKey` is the key name used by that codec.

### Example shapes + mappings

```
User : object
    + id        : string
    + name      : object
        + first : string
        + last  : string
    - email     : string
    - locale    : string
    + userAgent : string

OrderList : object[]
    + id       : string
    + amount   : float
    + currency : string
    - note     : string

codec json User
    id         = "id"
    name.first = "first_name"
    name.last  = "last_name"
    email      = "email"
    locale     = "locale"
    userAgent  = "User Agent"

codec query User
    id         = "user_id"
    name.first = "first"
    name.last  = "last"
    email      = "email"
    locale     = "lang"
    userAgent  = "user_agent"

codec yaml User
    id         = "id"
    name.first = "first"
    name.last  = "last"
    email      = "email"
    locale     = "locale"
    userAgent  = "user_agent"

codec csv OrderList
    id       = "id"
    amount   = "amount"
    currency = "currency"
    note     = "note"

codec tsv OrderList
    id       = "id"
    amount   = "amount"
    currency = "currency"
    note     = "note"
```

### Intended semantics (draft)

- `decode(text, codec, Shape)`:
  1) Decode text into a raw value.
  2) Apply the shape.
  3) Use codec mappings to translate external keys → internal fields.

- `encode(value, codec, Shape)`:
  1) Apply the shape.
  2) Use codec mappings to translate internal fields → external keys.
  3) Encode to the target format.

- If a path is missing in a codec block, use the **internal field name**.
- If a codec block is absent for a shape, use internal field names for all paths.

### End-to-end translation examples (draft)

Assume builtins (not implemented yet):
```
decode(text, codec, Shape?)
encode(value, codec, Shape?)
```

Inputs:

```
let userJson = "{ \"id\": \"u1\", \"first_name\": \"Ada\", \"last_name\": \"Lovelace\", \"email\": \"ada@karl.dev\", \"locale\": \"en\", \"User Agent\": \"Karl/1.0\" }"
let userYaml = "id: u1\nfirst: Ada\nlast: Lovelace\nemail: ada@karl.dev\nlocale: en\nuser_agent: Karl/1.0\n"
let ordersJson = "[{ \"id\": \"o1\", \"amount\": 10.5, \"currency\": \"USD\" }, { \"id\": \"o2\", \"amount\": 7.0, \"currency\": \"EUR\", \"note\": \"gift\" }]"
let ordersCsv = "id,amount,currency,note\no1,10.5,USD,\no2,7.0,EUR,gift\n"
let ordersTsv = "id\tamount\tcurrency\tnote\no1\t10.5\tUSD\t\no2\t7.0\tEUR\tgift\n"
```

Translations:

1) JSON → CSV (OrderList)
```
let orders = decode(ordersJson, "json", OrderList)
let csv = encode(orders, "csv", OrderList)
```

2) JSON → YAML (User)
```
let user = decode(userJson, "json", User)
let yaml = encode(user, "yaml", User)
```

3) JSON → Query (User)
```
let user = decode(userJson, "json", User)
let query = encode(user, "query", User)
```

4) YAML → JSON (User)
```
let user = decode(userYaml, "yaml", User)
let json = encode(user, "json", User)
```

5) CSV → YAML (OrderList)
```
let orders = decode(ordersCsv, "csv", OrderList)
let yaml = encode(orders, "yaml", OrderList)
```

6) TSV → CSV (OrderList)
```
let orders = decode(ordersTsv, "tsv", OrderList)
let csv = encode(orders, "csv", OrderList)
```

7) CSV → JSON (OrderList)
```
let orders = decode(ordersCsv, "csv", OrderList)
let json = encode(orders, "json", OrderList)
```

### Transform in between (draft)

Decoding yields a shaped object, so you can transform it with normal Karl code
before encoding into a different format.

Example (User → YAML):
```
let user = decode(userJson, "json", User)
user.locale = "fr"
user.name.last = user.name.last.toUpper()
let out = encode(user, "yaml", User)
```

Example (OrderList → JSON with discounts):
```
let orders = decode(ordersCsv, "csv", OrderList)
let discounted = for i < orders.length with i = 0, out = [] {
    let o = orders[i]
    o.amount = o.amount * 0.9
    out += [o]
    i++
} then out
let json = encode(discounted, "json", OrderList)
```

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
