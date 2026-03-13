// Read newline-delimited integers from stdin and keep the top 5.
// Example:
// echo -e "7\n2\n9\n1\n5\n8" | karl run examples/features/streams/stdin_numbers_top.k

let top5 = stdin({ type: TEXT, })
    | lines()
    | filter(_ != "")
    | map(parseInt(_))
    | top(5)

log(top5)
