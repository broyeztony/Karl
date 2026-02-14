#!/bin/bash
# Connect to remote Karl REPL with readline support

if [ $# -eq 0 ]; then
    echo "Usage: $0 <host:port>"
    echo ""
    echo "Examples:"
    echo "  $0 localhost:9000"
    echo "  $0 192.168.1.100:9000"
    exit 1
fi

SERVER="$1"

# Check if rlwrap is installed
if command -v rlwrap &> /dev/null; then
    echo "Connecting to $SERVER with rlwrap..."
    exec rlwrap -c -H ~/.karl_remote_history -r ./karl repl-client "$SERVER"
else
    echo "Connecting to $SERVER..."
    echo "(Install rlwrap for command history: brew install rlwrap)"
    exec ./karl repl-client "$SERVER"
fi
