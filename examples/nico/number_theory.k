// Number Theory Explorer
// Explore divisors, perfect numbers, and number properties

// Find all divisors of a number
let find_divisors = (n) -> {
    for i < n with i = 1, divisors = [] {
        if n % i == 0 {
            divisors += [i]
        }
        i++
    } then divisors
}

// Sum of divisors (excluding the number itself)
let sum_proper_divisors = (n) -> {
    let divs = find_divisors(n)
    for i < divs.length - 1 with i = 0, sum = 0 {
        sum += divs[i]
        i++
    } then sum
}

// Check if a number is perfect (sum of divisors equals the number)
let is_perfect = (n) -> {
    sum_proper_divisors(n) == n
}

// Check if a number is abundant (sum of divisors > number)
let is_abundant = (n) -> {
    sum_proper_divisors(n) > n
}

// Check if  a number is deficient (sum of divisors < number)
let is_deficient = (n) -> {
    sum_proper_divisors(n) < n
}

// Classify a number
let classify_number = (n) -> {
    let sum_divs = sum_proper_divisors(n)
    if sum_divs == n {
        "perfect"
    } else {
        if sum_divs > n {
            "abundant"
        } else {
            "deficient"
        }
    }
}

// Greatest Common Divisor (GCD) using Euclidean algorithm
let gcd = (a, b) -> {
    for continue_loop < 100 with continue_loop = 0, x = a, y = b {
        if y == 0 {
            break x
        }
        
        let temp = y
        y = x % y
        x = temp
        continue_loop++
    } then x
}

// Check if two numbers are coprime (GCD = 1)
let are_coprime = (a, b) -> {
    gcd(a, b) == 1
}

// Find all perfect numbers up to limit
let find_perfect_numbers = (limit) -> {
    for n < limit with n = 2, perfects = [] {
        if is_perfect(n) {
            perfects += [n]
        }
        n++
    } then perfects
}

log("🔢 Number Theory Explorer")
log("═════════════════════════")

// Analyze specific numbers
let test_numbers = [6, 12, 28, 30, 24, 18]

log("\n📊 Number Classifications:")
for i < test_numbers.length with i = 0 {
    let n = test_numbers[i]
    let divs = find_divisors(n)
    let sum = sum_proper_divisors(n)
    let classification = classify_number(n)
    
    log("\nNumber:", n)
    log("  Divisors:", divs)
    log("  Sum of proper divisors:", sum)
    log("  Classification:", classification)
    i++
} then {}

log("\n✨ Perfect Numbers up to 100:")
let perfects = find_perfect_numbers(101)
log(perfects, "- these are rare and special!")

log("\n🔗 GCD Examples:")
log("gcd(48, 18) =", gcd(48, 18))
log("gcd(17, 19) =", gcd(17, 19), "(coprime)")
log("gcd(100, 50) =", gcd(100, 50))

log("\n🎲 Coprime pairs in range 10-15:")
for a < 16 with a = 10 {
    for b < 16 with b = a + 1 {
        if are_coprime(a, b) {
            log("  (", a, ",", b, ") are coprime")
        }
        b++
    } then {}
    a++
} then {}

{
    perfect_numbers: perfects,
    classifications: test_numbers.map(n -> {
        number: n,
        type: classify_number(n),
        divisor_sum: sum_proper_divisors(n),
    }),
    gcd_examples: {
        gcd_48_18: gcd(48, 18),
        gcd_17_19: gcd(17, 19),
    },
}
