#!/bin/bash
# Karl REPL with readline support (command history, arrow keys, etc.)

# Check if rlwrap is installed
if ! command -v rlwrap &> /dev/null; then
    echo "rlwrap is not installed."
    echo ""
    echo "To install rlwrap:"
    echo "  macOS:   brew install rlwrap"
    echo "  Ubuntu:  sudo apt-get install rlwrap"
    echo "  Fedora:  sudo dnf install rlwrap"
    echo ""
    echo "Starting Karl REPL without rlwrap..."
    exec ./karl repl
fi

# Start Karl REPL with rlwrap
# -c: Enable filename completion
# -H: Use history file
# -r: Put all words on completion list
exec rlwrap -c -H ~/.karl_history -r ./karl repl
