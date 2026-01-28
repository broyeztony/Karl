# Nico's Karl Examples

Eight cool programs demonstrating different features of the Karl programming language.

## Programs

### 1. fibonacci_sequence.k
**Fibonacci Sequence Generator**

Demonstrates:
- Recursive functions
- Iterative loops with state management (`for` with `with` clause)
- List generation and manipulation
- Higher-order functions (`.filter()`, `.sum()`)
- Multiple implementation approaches to solve the same problem

Features:
- Generates the first 15 Fibonacci numbers
- Filters numbers under 200
- Calculates sum of even Fibonacci numbers
- Compares recursive vs iterative approaches

Run with:
```bash
cat examples/nico/fibonacci_sequence.k | ./karl run -
```

### 2. prime_sieve.k
**Prime Number Generator (Sieve of Eratosthenes)**

Demonstrates:
- Nested loop control flow
- Early breaking with `break`
- Complex conditionals with `if/else`
- Data categorization and filtering
- Array indexing and manipulation
- Object literal creation with shorthand syntax

Features:
- Finds all prime numbers under 100
- Categorizes primes by size (small/medium/large)
- Identifies twin primes (pairs with difference of 2)
- Comprehensive statistical analysis

Run with:
```bash
cat examples/nico/prime_sieve.k | ./karl run -
```

### 3. text_analyzer.k
**Text Analyzer - String Processing & Pattern Detection**

Demonstrates:
- String manipulation (`.toLower()`, `.split()`, `.chars()`)
- Map operations (`.set()`, `.get()`, `.has()`, `.keys()`)
- Palindrome detection algorithm
- Word frequency counting
- Pattern matching with `match`
- Complex data structure building

Features:
- Detects palindromes in text
- Counts word frequencies
- Finds repeated words
- Calculates text statistics
- Uses pattern matching for result reporting

Run with:
```bash
cat examples/nico/text_analyzer.k | ./karl run -
```

### 4. matrix_operations.k
**Linear Algebra & Matrix Operations**

Demonstrates:
- 2D array creation and manipulation
- Nested loops for matrix operations
- Higher-order functions with `.map()`
- Functional programming patterns
- Mathematical operations on complex structures

Features:
- Creates matrices with custom fill functions
- Identity matrix generation
- Matrix transpose
- Matrix addition
- Scalar multiplication
- Row sum calculations

Run with:
```bash
cat examples/nico/matrix_operations.k | ./karl run -
```

### 5. collatz_conjecture.k
**Collatz Conjecture Explorer (3n + 1 Problem)**

Demonstrates:
- Custom integer division implementation
- Long running sequences
- Peak finding algorithms
- Sorting and analysis
- Famous unsolved mathematical problem

Features:
- Generates Collatz sequences for any number
- Finds peak values in sequences
- Identifies longest sequences in ranges
- Demonstrates the famous 27 → 9232 peak case
- Analyzes and ranks multiple starting numbers

Run with:
```bash
cat examples/nico/collatz_conjecture.k | ./karl run -
```

### 6. sorting_algorithms.k
**Sorting Algorithm Comparison**

Demonstrates:
- Classic sorting algorithms (Bubble, Selection, Insertion)
- Nested loops with complex logic
- Array mutation and swapping
- Algorithm comparison
- Validation functions

Features:
- Implements three classic sorting algorithms
- Compares results with Karl's built-in sort
- Handles various input types (reversed, duplicates)
- Validates sorted arrays
- Educational comparison of approaches

Run with:
```bash
cat examples/nico/sorting_algorithms.k | ./karl run -
```

### 7. pascal_triangle.k
**Pascal's Triangle Generator**

Demonstrates:
- Recursive row generation
- Mathematical patterns and properties
- Multi-dimensional array analysis
- Position finding in nested structures
- Symmetry checking

Features:
- Generates Pascal's Triangle to any depth
- Verifies row sums are powers of 2
- Finds specific values in the triangle
- Checks symmetry properties
- Beautiful mathematical visualization

Run with:
```bash
cat examples/nico/pascal_triangle.k | ./karl run -
```

### 8. number_theory.k
**Number Theory Explorer**

Demonstrates:
- Divisor finding algorithms
- Perfect number detection
- GCD using Euclidean algorithm
- Number classification (perfect, abundant, deficient)
- Mathematical property checking

Features:
- Finds all divisors of numbers
- Identifies perfect numbers (sum of divisors equals the number)
- Calculates GCD of two numbers
- Finds coprime pairs
- Classifies numbers by their divisor properties

Run with:
```bash
cat examples/nico/number_theory.k | ./karl run -
```

## Key Karl Features Demonstrated

### Control Flow
- `for` loops with `with` clause for loop-local state
- `break` and `continue` statements with values
- `then` clause for loop result expressions
- `if/else` expressions
- `match` pattern matching with ranges and wildcards

### Functions
- Arrow function syntax `() ->`
- Closures and higher-order functions
- Recursive functions
- Custom helper functions for complex operations

### Data Structures
- Arrays/Lists with methods: `.filter()`, `.map()`, `.sum()`, `.length`, `.sort()`
- Maps with `.set()`, `.get()`, `.has()`, `.keys()`, `.values()`
- Sets with `.add()`, `.has()`, `.delete()`, `.size`
- Object literals with shorthand and spread syntax
- 2D arrays (matrices)

### String Processing
- String methods: `.toLower()`, `.toUpper()`, `.split()`, `.chars()`
- Character-by-character processing
- Pattern detection and analysis

### Mathematical Operations
- Custom integer arithmetic to avoid float conversion
- Modulo operator for divisibility
- Mathematical algorithms (GCD, divisors, primes)
- Sequence generation and analysis

### Advanced Patterns
- Loop-based iteration with state management
- Custom comparison functions for sorting
- Functional transformations with map/filter/reduce
- Complex object construction

---

**Author:** Nico  
**Language:** Karl  
**Created:** 2026-01-28  
**Total Programs:** 8  
**Lines of Code:** ~600+
