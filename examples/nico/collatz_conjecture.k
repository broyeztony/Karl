// Collatz Conjecture Explorer (3n + 1 Problem)
// One of math's most famous unsolved problems!
// Rule: If n is even, divide by 2. If odd, multiply by 3 and add 1.
// Conjecture: Every positive integer eventually reaches 1.

// Simple manual division by 2 for integers
let half = (n) -> {
    for i < n with i = 0, result = 0 {
        if i * 2 >= n {
            break result
        }
        result = result + 1
        i++
    } then result
}

// Generate the Collatz sequence for a number
let collatz_sequence = (n) -> {
    for continue_loop < 200 with continue_loop = 0, current = n, sequence = [n] {
        if current == 1 {
            break sequence
        }
        
        if current % 2 == 0 {
            current = half(current)
        } else {
            current = current * 3 + 1
        }
        
        sequence += [current]
        continue_loop++
    } then sequence
}

// Find the maximum value reached in the sequence (peak height)
let peak_height = (n) -> {
    let seq = collatz_sequence(n)
    for i < seq.length with i = 0, max_val = n {
        if seq[i] > max_val {
            max_val = seq[i]
        }
        i++
    } then max_val
}

// Analyze a range of starting numbers
let analyze_range = (start, end) -> {
    for n < end with n = start, results = [] {
        let seq = collatz_sequence(n)
        results += [{
            number: n,
            steps: seq.length - 1,
            peak: peak_height(n),
        }]
        n++
    } then results
}

// Find the number with the longest sequence in a range
let find_longest_sequence = (start, end) -> {
    for n < end with n = start, longest = { number: start, steps: 0 } {
        let steps = collatz_sequence(n).length - 1
        if steps > longest.steps {
            longest = { number: n, steps: steps, }
        }
        n++
    } then longest
}

// Explore some interesting numbers
let num_6 = collatz_sequence(6)    // Simple example: 6 → 3 → 10 → 5 → 16 → 8 → 4 → 2 → 1
let num_27 = collatz_sequence(27)  // Famous for its long sequence and high peak

// Analyze range 1-20
let analysis = analyze_range(1, 21)

// Find longest in range 1-30
let champion = find_longest_sequence(1, 31)

log("🔢 Collatz Conjecture Explorer")
log("══════════════════════════════")
log("\n📊 Example: Starting with 6")
log("Sequence:", num_6)
log("Steps to reach 1:", num_6.length - 1)
log("Peak value:", peak_height(6))

log("\n📊 Example: Starting with 27")
log("Sequence:", num_27)
log("Steps to reach 1:", num_27.length - 1)
log("Peak value:", peak_height(27))

log("\n🏆 Champion in range 1-30:")
log("Number:", champion.number)
log("Steps:", champion.steps)

log("\n📈 Top 5 sequences by length in range 1-20:")
let sorted_analysis = analysis.sort((a, b) -> b.steps - a.steps)
for i < 5 with i = 0 {
    let a = sorted_analysis[i]
    log("  n=", a.number, "→ steps=", a.steps, "peak=", a.peak)
    i++
} then {}

{
    example_6: {
        sequence: num_6,
        steps: num_6.length - 1,
    },
    example_27: {
        sequence: num_27,
        steps: num_27.length - 1,
        peak: peak_height(27),
    },
    champion_1_to_30: champion,
    analysis: sorted_analysis,
}
