# Karl Notebook System

A minimal notebook system for Karl, similar to Jupyter notebooks. Execute Karl code in sequential cells while maintaining state across cells.

## Quick Start

### Run a notebook

```bash
karl notebook example.knb
```

### Save execution results

```bash
karl notebook example.knb --output=results.json
```

## Notebook Format

Notebooks are stored as JSON files with the `.knb` extension. A notebook contains:

- **title**: The notebook title (string)
- **description**: Optional description (string)
- **version**: Notebook format version (string, default: "1.0")
- **cells**: Array of cell objects
- **metadata**: Optional metadata (object)

### Cell Format

Each cell has:

```json
{
  "type": "code" | "markdown",
  "title": "Optional cell title",
  "source": "Cell content",
  "metadata": {}
}
```

#### Code Cells

Code cells execute Karl code and produce outputs:

```json
{
  "type": "code",
  "title": "Calculate sum",
  "source": "1 + 2 + 3"
}
```

#### Markdown Cells

Markdown cells contain documentation and are not executed:

```json
{
  "type": "markdown",
  "source": "# Section Header\nSome documentation text"
}
```

## Minimal Notebook Example

```json
{
  "title": "My First Notebook",
  "description": "A simple example",
  "version": "1.0",
  "cells": [
    {
      "type": "markdown",
      "source": "# Hello Karl\nThis is a notebook."
    },
    {
      "type": "code",
      "title": "Arithmetic",
      "source": "2 + 3"
    },
    {
      "type": "code",
      "title": "Variables",
      "source": "let x = 10\nlet y = 20\nx + y"
    }
  ]
}
```

## State Persistence

Variables and functions defined in one code cell persist to subsequent cells:

```json
{
  "cells": [
    {
      "type": "code",
      "source": "let double = x -> x * 2"
    },
    {
      "type": "code",
      "source": "double(21)"  // Uses function from previous cell
    }
  ]
}
```

## Output Format

When executing notebooks, outputs are displayed in order:

```
Notebook: My Notebook
Executed 5 cells

Cell 0: 5
Cell 1: Hello
Cell 2: [1, 2, 3]
Cell 3 [ERROR]: Parse error: unexpected token
```

### Saving Results with `--output`

Results are saved as a JSON file containing both the notebook definition and execution outputs:

```json
{
  "notebook": { ... },
  "outputs": [
    {
      "cell_index": 0,
      "type": "result",
      "value": "5",
      "timestamp": "2024-01-15T10:30:00Z"
    },
    ...
  ]
}
```

## Examples

Four example notebooks are provided in `notebook/examples/`:

1. **01-quickstart.knb** - Basic operations, variables, lists, objects
2. **02-functions-closures.knb** - Functions, closures, composition, higher-order functions
3. **03-collections.knb** - Lists, sets, maps, pattern matching with ranges
4. **04-advanced.knb** - Match expressions, guards, recover operator, destructuring

### Run an example

```bash
karl notebook notebook/examples/01-quickstart.knb
```

## Usage Patterns

### Educational Walkthroughs

Use markdown cells to document concepts alongside code:

```json
{
  "type": "markdown",
  "source": "## Closures\nClosures allow functions to capture variables from their scope."
},
{
  "type": "code",
  "title": "Example: makeAdder closure",
  "source": "let makeAdder = n -> x -> n + x\nlet add5 = makeAdder(5)\nadd5(10)"
}
```

### Data Analysis

Keep state across multiple cells for step-by-step analysis:

```json
{
  "type": "code",
  "title": "Load data",
  "source": "let data = loadData(\"file.json\")"
},
{
  "type": "code",
  "title": "Transform",
  "source": "let processed = transform(data)"
},
{
  "type": "code",
  "title": "Analyze",
  "source": "analyze(processed)"
}
```

### Testing & Debugging

Use notebooks to explore code behavior interactively:

```json
{
  "type": "code",
  "source": "let myFunc = x -> x * 2\nmyFunc(5)"
},
{
  "type": "code",
  "source": "myFunc(100)"
}
```

## Architecture

The notebook system consists of:

- **notebook.go**: Core data structures and runner
  - `Notebook`: Represents a complete notebook
  - `Cell`: Individual cells (code/markdown)
  - `Runner`: Executes cells with persistent environment
  - `CellOutput`: Execution results

- **main.go**: CLI integration
  - `notebookCommand()`: Entry point for `karl notebook` command
  - `notebookUsage()`: Help text

## Programmatic Usage

Use the notebook package directly in Go code:

```go
package main

import "karl/notebook"

func main() {
    // Load notebook
    nb, _ := notebook.LoadNotebook("example.knb")
    
    // Create runner
    runner := notebook.NewRunner()
    
    // Execute all cells
    outputs, _ := runner.ExecuteNotebook(nb)
    
    // Process outputs
    for _, output := range outputs {
        println(output.Value)
    }
}
```

## Limitations & Future Enhancements

Current minimal implementation:

- No interactive editing (static JSON files only)
- No visualization of results
- Output results are text-only
- No cell dependencies or conditional execution

Potential future enhancements:

- Interactive notebook editor/server
- Rich output support (tables, charts, images)
- Cell tagging and organization
- Dependency tracking between cells
- Notebook versioning and history
- Export to HTML/PDF
