// Sorting Algorithm Comparison
// Compare different sorting approaches and their characteristics

// Bubble Sort - Simple but inefficient
let bubble_sort = (arr) -> {
    let n = arr.length
    let sorted = arr
    
    for i < n with i = 0 {
        for j < n - i - 1 with j = 0 {
            if sorted[j] > sorted[j + 1] {
                // Swap elements
                let temp = sorted[j]
                sorted[j] = sorted[j + 1]
                sorted[j + 1] = temp
            }
            j++
        } then {}
        i++
    } then sorted
}

// Selection Sort - Find minimum and place it
let selection_sort = (arr) -> {
    let n = arr.length
    let sorted = arr
    
    for i < n - 1 with i = 0 {
        let min_idx = i
        
        for j < n with j = i + 1 {
            if sorted[j] < sorted[min_idx] {
                min_idx = j
            }
            j++
        } then {}
        
        // Swap
        if min_idx != i {
            let temp = sorted[i]
            sorted[i] = sorted[min_idx]
            sorted[min_idx] = temp
        }
        
        i++
    } then sorted
}

// Insertion Sort - Build sorted array one element at a time
let insertion_sort = (arr) -> {
    let n = arr.length
    let sorted = arr
    
    for i < n with i = 1 {
        let key = sorted[i]
        let j = i - 1
        
        // Move elements greater than key one position ahead
        for continue_loop < n with continue_loop = 0 {
            if j >= 0 && sorted[j] > key {
                sorted[j + 1] = sorted[j]
                j = j - 1
                continue_loop++
            } else {
                break {}
            }
        } then {}
        
        sorted[j + 1] = key
        i++
    } then sorted
}

// Quick check if array is sorted
let is_sorted = (arr) -> {
    for i < arr.length - 1 with i = 0, sorted = true {
        if arr[i] > arr[i + 1] {
            sorted = false
            break sorted
        }
        i++
    } then sorted
}

// Test data
let test_small = [64, 34, 25, 12, 22, 11, 90]
let test_reversed = [10, 9, 8, 7, 6, 5, 4, 3, 2, 1]
let test_partial = [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]

log("🔀 Sorting Algorithm Comparison")
log("═══════════════════════════════")

log("\n📊 Original array:", test_small)
log("Bubble Sort:   ", bubble_sort(test_small))
log("Selection Sort:", selection_sort(test_small))
log("Insertion Sort:", insertion_sort(test_small))

log("\n📊 Reversed array:", test_reversed)
log("Bubble Sort:   ", bubble_sort(test_reversed))

log("\n📊 With duplicates:", test_partial)
log("Insertion Sort:", insertion_sort(test_partial))

log("\n✅ Verification:")
log("Is [1,2,3] sorted?", is_sorted([1, 2, 3]))
log("Is [3,1,2] sorted?", is_sorted([3, 1, 2]))

// Compare with Karl's built-in sort
let builtin_result = test_small.sort((a, b) -> a - b)

{
    original: test_small,
    bubble_sorted: bubble_sort(test_small),
    selection_sorted: selection_sort(test_small),
    insertion_sorted: insertion_sort(test_small),
    builtin_sorted: builtin_result,
    all_match: bubble_sort(test_small) == builtin_result,
}
