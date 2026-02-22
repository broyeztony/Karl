// Data transformation with grouping and aggregation patterns.

let sales = [
    { product: "apple", qty: 5, price: 2.0 },
    { product: "banana", qty: 3, price: 1.5 },
    { product: "apple", qty: 2, price: 2.0 },
    { product: "orange", qty: 4, price: 3.0 },
    { product: "banana", qty: 1, price: 1.5 },
]

// Transform: add total per sale
let withTotal = sales.map(s -> { product: s.product, qty: s.qty, price: s.price, total: s.qty * s.price, })

// Group by product and sum quantities
let grouped = for i < withTotal.length with i = 0, acc = map() {
    let item = withTotal[i]
    let current = if acc.has(item.product) {
        acc.get(item.product)
    } else {
        { qty: 0, revenue: 0.0 }
    }
    let updated = { qty: current.qty + item.qty, revenue: current.revenue + item.total, }
    acc.set(item.product, updated)
    i++
} then acc

// Convert map values to array and encode as JSON
let reportArray = grouped.values()
let report = jsonEncode(reportArray)

report
