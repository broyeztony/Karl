#!/bin/bash
# Comprehensive test of all REPL examples

echo "=== Testing All REPL Examples ==="
echo ""

echo "Test: Basic Arithmetic and Variables"
printf "let x = 10\nlet y = 20\nx + y\nx * y\n:quit\n" | ./karl repl | grep -q "30" && echo "✅ PASS" || echo "❌ FAIL"

echo "Test: First-Class Functions"
printf "let double = x -> x * 2\nlet inc = x -> x + 1\nlet compose = (f, g) -> x -> f(g(x))\nlet doubleAndInc = compose(inc, double)\ndoubleAndInc(20)\n:quit\n" | ./karl repl | grep -q "41" && echo "✅ PASS" || echo "❌ FAIL"

echo "Test: Closures"
printf "let makeAdder = n -> x -> n + x\nlet add10 = makeAdder(10)\nadd10(5)\nadd10(32)\n:quit\n" | ./karl repl | grep -q "42" && echo "✅ PASS" || echo "❌ FAIL"

echo "Test: Lists"
printf "let nums = [1, 2, 3, 4, 5]\nnums.length\nnums[0]\nnums[2]\n:quit\n" | ./karl repl | grep -q "5" && echo "✅ PASS" || echo "❌ FAIL"

echo "Test: Objects"
printf 'let person = { name: "Alice", age: 30 }\nperson.name\nperson.age\n:quit\n' | ./karl repl | grep -q "Alice" && echo "✅ PASS" || echo "❌ FAIL"

echo "Test: Match expressions"
printf 'let x = 5\nmatch x { case 5 -> "five" case _ -> "other" }\n:quit\n' | ./karl repl | grep -q "five" && echo "✅ PASS" || echo "❌ FAIL"

echo ""
echo "=== All Example Tests Complete ==="
