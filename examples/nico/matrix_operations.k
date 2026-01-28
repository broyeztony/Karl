// Matrix Operations - Linear Algebra in Karl
// Demonstrates 2D array manipulation, nested loops, and mathematical operations

// Create a matrix (2D array)
let create_matrix = (rows, cols, fill_fn) -> {
    for i < rows with i = 0, matrix = [] {
        let row = for j < cols with j = 0, r = [] {
            r += [fill_fn(i, j)]
            j++
        } then r
        matrix += [row]
        i++
    } then matrix
}

// Create an identity matrix
let identity = (size) -> {
    create_matrix(size, size, (i, j) -> if i == j { 1 } else { 0 })
}

// Transpose a matrix
let transpose = (matrix) -> {
    let rows = matrix.length
    let cols = matrix[0].length
    
    for i < cols with i = 0, result = [] {
        let row = for j < rows with j = 0, r = [] {
            r += [matrix[j][i]]
            j++
        } then r
        result += [row]
        i++
    } then result
}

// Matrix addition
let matrix_add = (a, b) -> {
    let rows = a.length
    let cols = a[0].length
    
    for i < rows with i = 0, result = [] {
        let row = for j < cols with j = 0, r = [] {
            r += [a[i][j] + b[i][j]]
            j++
        } then r
        result += [row]
        i++
    } then result
}

// Scalar multiplication
let scalar_multiply = (matrix, scalar) -> {
    matrix.map(row -> row.map(val -> val * scalar))
}

// Get row sum
let row_sum = (matrix, row_idx) -> {
    matrix[row_idx].sum()
}

// Print matrix (for visualization)
let matrix_to_string = (matrix) -> {
    for i < matrix.length with i = 0, lines = [] {
        let row_str = for j < matrix[i].length with j = 0, parts = [] {
            parts += [matrix[i][j]]
            j++
        } then parts
        lines += [row_str]
        i++
    } then lines
}

// Create test matrices
let mat_a = create_matrix(3, 3, (i, j) -> i * 3 + j + 1)
let mat_b = create_matrix(3, 3, (i, j) -> (i + j) * 2)
let id3 = identity(3)

// Perform operations
let mat_sum = matrix_add(mat_a, mat_b)
let mat_scaled = scalar_multiply(mat_a, 2)
let mat_transposed = transpose(mat_a)

log("🧮 Matrix Operations Demo")
log("\nMatrix A (3x3):", matrix_to_string(mat_a))
log("\nMatrix B (3x3):", matrix_to_string(mat_b))
log("\nIdentity Matrix (3x3):", matrix_to_string(id3))
log("\nA + B =", matrix_to_string(mat_sum))
log("\n2 * A =", matrix_to_string(mat_scaled))
log("\nA^T (transposed) =", matrix_to_string(mat_transposed))
log("\nRow 1 sum of A:", row_sum(mat_a, 1))

{
    matrix_a: mat_a,
    matrix_b: mat_b,
    identity: id3,
    sum: mat_sum,
    scaled: mat_scaled,
    transposed: mat_transposed,
}
