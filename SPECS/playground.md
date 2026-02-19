# Karl Playground Specification

## Overview
The Karl Playground is a web-based environment for writing and executing Karl code directly in the browser. It leverages WebAssembly (WASM) to run the Karl interpreter client-side, eliminating the need for a backend execution service.

## Architecture
- **Language**: Go (Karl Interpreter) compiled to WASM (`GOOS=js GOARCH=wasm`).
- **Frontend**: HTML5 + Vanilla JS.
- **Runtime**: `wasm_exec.js` (Go's JS glue code).
- **Communication**: 
  - Compilation: `cmd/wasm/main.go` exposes `runKarl(source)` to the global JS scope.
  - I/O: JS overrides `window.fs.writeSync` to capture stdout/stderr from the Go runtime and display it in the output pane.

## Usage

### Starting the Server
Run the playground server locally:
```bash
go run main.go playground
# or
./karl playground
```
Open [http://localhost:8081](http://localhost:8081) in your browser.

### Building the WASM Binary
If you modify the interpreter or core logic, you must rebuild the WASM binary:
```bash
GOOS=js GOARCH=wasm go build -o assets/playground/karl.wasm ./wasm
```

### Dependencies
- `assets/playground/wasm_exec.js`: Must match the Go version used to compile the WASM binary.
- `assets/playground/karl.wasm`: The compiled interpreter.
- `assets/playground/index.html`: The UI.

## Parse Error UX (Before/After)

### Input example
```karl
// Welcome to Karl Playground
// Karl is a functional-first language.

// Define a function
let add = (a, b) -> a + b

// Use list operations
let nums = [1, 2, 3, 4, 5]
let squares = nums.map((x) -> x * x)

log("Squares:", squares)
sleep(2000) // Sleep for 2 seconds (non-blocking in worker)
log("Sum of squares:", squares.reduce(add, 0))

// Example of concurrency
// Example of concurrency
let c = channel()
let spawn = (ch) -> {
    ch.send("Hello from thread!")
    ch.done()
}

& spawn(c)
log("Received:" c.recv())
log("Received:" c.recv())
```

### Before
```text
Parse Errors:
expected next token to be ), got IDENT instead
no prefix parse function for ) found
expected next token to be ), got IDENT instead
no prefix parse function for ) found
```

### After
```text
parse error: expected next token to be ), got IDENT instead
  at playground.k:24:17
  24 | log("Received:" c.recv())
    |                 ^
parse error: no prefix parse function for ) found
  at playground.k:24:25
  24 | log("Received:" c.recv())
    |                         ^
parse error: expected next token to be ), got IDENT instead
  at playground.k:25:17
  25 | log("Received:" c.recv())
    |                 ^
parse error: no prefix parse function for ) found
  at playground.k:25:25
  25 | log("Received:" c.recv())
    |                         ^
```

## Limitations
- **File System**: The browser environment does not have direct file system access. `readFile`/`writeFile` calls will fail unless polyfilled.
- **Network**: Direct socket access is restricted by browser security policies.
- **Performance**: Large programs run on the main thread and may impact UI responsiveness.
