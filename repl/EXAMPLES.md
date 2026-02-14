# Karl REPL Examples

Try these examples in the Karl REPL (`./karl repl`):

## 1. Basic Arithmetic and Variables

```karl
let x = 10
let y = 20
x + y
x * y
```

## 2. First-Class Functions

```karl
let double = x -> x * 2
let inc = x -> x + 1
let compose = (f, g) -> x -> f(g(x))
let doubleAndInc = compose(inc, double)
doubleAndInc(20)  // Result: 41
```

## 3. Closures

```karl
let makeCounter = () -> {
  let count = 0
  () -> {
    count = count + 1
    count
  }
}
let counter = makeCounter()
counter()  // 1
counter()  // 2
counter()  // 3
```

## 4. Higher-Order Functions

```karl
let map = (fn, list) -> {
  let result = []
  for i < list.length with i = 0 {
    result = result + [fn(list[i])]
    i++
  } then result
}
let nums = [1, 2, 3, 4, 5]
map(x -> x * 2, nums)  // [2, 4, 6, 8, 10]
```

## 5. Pattern Matching

```karl
let describe = n -> match n {
  case 0 -> "zero"
  case 1 -> "one"
  case _ if n < 0 -> "negative"
  case _ if n > 100 -> "large"
  case _ -> "other"
}
describe(0)    // "zero"
describe(-5)   // "negative"
describe(150)  // "large"
```

## 6. Objects and Destructuring

```karl
let person = {
  name: "Alice",
  age: 30,
  city: "Paris"
}
let { name, age } = person
name  // "Alice"
age   // 30
```

## 7. List Operations

```karl
let nums = [1, 2, 3, 4, 5]
nums.length      // 5
nums[0]          // 1
nums[2]          // 3
[1, 2] + [3, 4]  // [1, 2, 3, 4]
```

## 8. String Manipulation

```karl
let text = "  Hello, Karl!  "
let trimmed = text.trim()
let lower = trimmed.toLower()
let upper = trimmed.toUpper()
let parts = trimmed.split(", ")
parts  // ["Hello", "Karl!"]
```

## 9. Recursive Functions

```karl
let factorial = n -> {
  if n <= 1 {
    1
  } else {
    n * factorial(n - 1)
  }
}
factorial(5)   // 120
factorial(10)  // 3628800
```

## 10. Loop as Expression

```karl
let findFirst = (list, pred) -> {
  for i < list.length with i = 0 {
    if pred(list[i]) {
      break list[i]
    }
    i++
  } then "not found"
}
let nums = [1, 3, 5, 8, 9, 11]
findFirst(nums, x -> x % 2 == 0)  // 8
```

## Tips

- Variables and functions persist across REPL evaluations
- Use `:help` to see available REPL commands
- Use `:quit` or Ctrl+C to exit
- Multi-line input is automatically detected
