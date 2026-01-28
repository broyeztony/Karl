// Pascal's Triangle Generator
// Beautiful mathematical pattern with many interesting properties

// Generate a single row of Pascal's triangle
let pascal_row = (n) -> {
    if n == 0 {
        [1]
    } else {
        let prev = pascal_row(n - 1)
        let row = [1]
        
        for i < prev.length - 1 with i = 0, new_row = row {
            new_row += [prev[i] + prev[i + 1]]
            i++
        } then new_row += [1]
    }
}

// Generate full Pascal's triangle up to n rows
let pascal_triangle = (rows) -> {
    for i < rows with i = 0, triangle = [] {
        triangle += [pascal_row(i)]
        i++
    } then triangle
}

// Sum of a row (always equals 2^n)
let row_sum = (row) -> row.sum()

// Find maximum value in a row
let row_max = (row) -> {
    for i < row.length with i = 0, max_val = 0 {
        if row[i] > max_val {
            max_val = row[i]
        }
        i++
    } then max_val
}

// Check if row is symmetric (it always is!)
let is_symmetric = (row) -> {
    let len = row.length
    for i < len / 2 with i = 0, symmetric = true {
        if row[i] != row[len - 1 - i] {
            symmetric = false
            break symmetric
        }
        i++
    } then symmetric
}

// Find all values in triangle that match a target
let find_value_positions = (triangle, target) -> {
    for i < triangle.length with i = 0, positions = [] {
        for j < triangle[i].length with j = 0 {
            if triangle[i][j] == target {
                positions += [{ row: i, col: j, }]
            }
            j++
        } then {}
        i++
    } then positions
}

// Format a row for display with spacing
let format_row = (row, max_width) -> {
    for i < row.length with i = 0, parts = [] {
        parts += [row[i]]
        i++
    } then parts
}

// Generate triangle
let triangle = pascal_triangle(10)

log("🔺 Pascal's Triangle (10 rows)")
log("═══════════════════════════════")

// Display the triangle
for i < triangle.length with i = 0 {
    log("Row", i, ":", format_row(triangle[i], 4))
    i++
} then {}

log("\n📊 Properties Analysis:")
log("Row 5 sum:", row_sum(triangle[5]), "(should be 2^5 = 32)")
log("Row 7 sum:", row_sum(triangle[7]), "(should be 2^7 = 128)")
log("Row 9 max value:", row_max(triangle[9]))
log("Row 6 is symmetric?", is_symmetric(triangle[6]))

log("\n🔍 Finding value 6 in triangle:")
let positions_6 = find_value_positions(triangle, 6)
log("Found at positions:", positions_6)

log("\n🔍 Finding value 20:")
let positions_20 = find_value_positions(triangle, 20)
log("Found at positions:", positions_20)

// Interesting property: Fibonacci sequence appears in diagonal sums!
log("\n✨ Fun fact: Diagonal sums form Fibonacci numbers!")
log("Try summing along diagonals to see the pattern!")

{
    triangle: triangle,
    row_5: triangle[5],
    row_5_sum: row_sum(triangle[5]),
    positions_of_6: positions_6,
    properties: {
        all_symmetric: true,
        powers_of_2: [
            row_sum(triangle[0]),
            row_sum(triangle[1]),
            row_sum(triangle[2]),
            row_sum(triangle[3]),
        ],
    },
}
