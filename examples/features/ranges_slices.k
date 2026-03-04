// Ranges build arrays; slices are half-open.

let nums = 1..5
let odds = 1..9 step 2
let letters = 'a'..'d'
let slice = nums[1..4]
let dropLast = nums[..-1]
let word = "some string"
let head = word[0..4]
let tail = word[5..]
let output = { nums, odds, letters, slice, dropLast, }

encodeJson(output)
