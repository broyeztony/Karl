#!/bin/bash
# Better test script for Karl REPL

echo "=== Karl REPL Demo ==="
echo ""

echo "Test 1: Simple arithmetic"
echo -e "1 + 2\n5 * 8\n100 / 4\n:quit" | ./karl repl
echo ""

echo "Test 2: Variable persistence across evaluations"
echo -e "let x = 10\nlet y = 20\nx\ny\nx + y\nx * y\n:quit" | ./karl repl
echo ""

echo "Test 3: String operations"
echo -e 'let name = "Karl"\nlet greeting = "Hello, " + name\ngreeting\n:quit' | ./karl repl
echo ""

echo "Test 4: Functions"
echo -e "let double = x -> x * 2\nlet inc = x -> x + 1\ndouble(21)\ninc(41)\n:quit" | ./karl repl
echo ""

echo "Test 5: Lists and operations"
echo -e "let nums = [1, 2, 3, 4, 5]\nnums.length\nnums[0]\nnums[2]\n:quit" | ./karl repl
echo ""

echo "Test 6: Objects"
echo -e 'let person = { name: "Alice", age: 30 }\nperson.name\nperson.age\n:quit' | ./karl repl
echo ""

echo "Test 7: Closures"
echo -e "let makeAdder = n -> x -> n + x\nlet add10 = makeAdder(10)\nadd10(5)\nadd10(32)\n:quit" | ./karl repl
echo ""

echo "Test 8: Match expressions"
echo -e 'let x = 5\nmatch x { case 5 -> "five" case _ -> "other" }\n:quit' | ./karl repl
echo ""

echo "=== All tests completed! ==="
