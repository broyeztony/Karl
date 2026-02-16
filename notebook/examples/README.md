# Karl Notebook Examples

This directory contains example notebooks for the Karl notebook system.

## Examples

### 01-quickstart.knb

**Level**: Beginner

A minimal introduction to Karl:
- Basic arithmetic operations
- String concatenation
- Variable binding
- Simple functions
- List operations
- Object/map operations

**Try it:**
```bash
karl notebook 01-quickstart.knb
```

### 02-functions-closures.knb

**Level**: Intermediate

Functional programming concepts:
- First-class functions
- Closures (`makeAdder` pattern)
- Function composition
- Higher-order functions (mapping)
- Reduce/accumulation patterns

**Key concepts:**
- Functions as values
- Lexical scoping
- Creating specialized functions dynamically

### 03-collections.knb

**Level**: Intermediate

Working with data structures:
- List creation and methods
- Sets for unique elements
- Maps/objects
- Pattern matching with ranges
- Filtering with predicates

**Key concepts:**
- Built-in collection types
- Set operations
- Pattern-based filtering

### 04-advanced.knb

**Level**: Advanced

Karl's distinctive features:
- Pattern matching syntax
- Match guards
- For loops as expressions
- Destructuring
- Optional chaining with `recover` operator

**Key concepts:**
- Expression-based control flow
- Error handling with recover
- Advanced pattern matching
- Accumulator patterns in loops

## Running Examples

### Run a single example
```bash
cd /workspaces/Karl
karl notebook notebook/examples/01-quickstart.knb
```

### Save results to a file
```bash
karl notebook notebook/examples/02-functions-closures.knb --output=results.json
```

### Run all examples
```bash
for file in notebook/examples/*.knb; do
  echo "=== Running $file ==="
  karl notebook "$file"
  echo
done
```

## Creating Your Own Notebooks

1. Create a new `.knb` file with the JSON notebook format
2. Add markdown cells for documentation
3. Add code cells with Karl code
4. Execute with `karl notebook mynotebook.knb`

Example template:

```json
{
  "title": "My Notebook",
  "description": "What this notebook demonstrates",
  "version": "1.0",
  "cells": [
    {
      "type": "markdown",
      "source": "# Introduction\nExplain what we're learning..."
    },
    {
      "type": "code",
      "title": "First example",
      "source": "let x = 10\nx * 2"
    },
    {
      "type": "code",
      "title": "Second example",
      "source": "let double = n -> n * 2\ndouble(x)"
    }
  ]
}
```

## Tips

- Start with 01-quickstart if you're new to Karl
- Use meaningful titles for code cells to document intent
- Combine markdown and code cells for a narrative flow
- Keep notebooks focused on a single topic
- Use the environment persistence to build complex examples gradually

## Next Steps

- Read [../README.md](../README.md) for complete documentation
- Explore the Karl language examples in `/examples/`
- Try the interactive REPL: `karl loom`
- Run the full test suite: `go test ./...`
