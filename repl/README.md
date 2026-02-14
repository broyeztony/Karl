# Karl REPL

An interactive Read-Eval-Print Loop for the Karl programming language.

## Usage

Start the REPL:

```bash
./karl repl
```

Or if you have `karl` in your PATH:

```bash
karl repl
```

## Features

- **Interactive evaluation**: Type expressions and see results immediately
- **Persistent environment**: Variables and functions persist across evaluations
- **Multi-line input**: Automatically detects incomplete expressions and prompts for continuation
- **REPL commands**: Built-in commands for help, environment inspection, and more

## REPL Commands

All commands start with a colon (`:`):

- `:help`, `:h` - Show help message
- `:quit`, `:q`, `:exit` - Exit the REPL
- `:env` - Show current environment bindings (coming soon)
- `:clear` - Clear the screen

## Examples

### Simple Expressions

```
karl> 1 + 2
3
karl> 5 * 8
40
karl> "Hello, " + "World"
"Hello, World"
```

### Variables

Variables persist across evaluations:

```
karl> let x = 10
karl> let y = 20
karl> x + y
30
karl> x * y
200
```

### Functions

Define and use functions:

```
karl> let double = x -> x * 2
karl> double(21)
42
karl> let add = (a, b) -> a + b
karl> add(10, 32)
42
```

### Closures

```
karl> let makeAdder = n -> x -> n + x
karl> let add10 = makeAdder(10)
karl> add10(5)
15
karl> add10(32)
42
```

### Lists and Objects

```
karl> let nums = [1, 2, 3, 4, 5]
karl> nums.length
5
karl> nums[0]
1

karl> let person = { name: "Alice", age: 30 }
karl> person.name
"Alice"
karl> person.age
30
```

### Match Expressions

```
karl> let x = 5
karl> match x { case 5 -> "five" case _ -> "other" }
"five"
```

## Multi-line Input

The REPL automatically detects incomplete input and prompts for continuation:

```
karl> let factorial = n -> {
...     if n <= 1 { 1 } else { n * factorial(n - 1) }
...   }
karl> factorial(5)
120
```

## Tips

- Press `Ctrl+C` or type `:quit` to exit
- The last expression's value is automatically printed
- Use `:help` to see available commands
- Variables and functions persist throughout your session

## Testing

Run the test suite:

```bash
./test_repl.sh
```

This runs a comprehensive set of tests demonstrating various Karl language features in the REPL.
