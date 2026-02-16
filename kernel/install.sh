#!/bin/bash
set -e

# Get the absolute path to the project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
KERNEL_DIR="$HOME/Library/Jupyter/kernels/karl"

echo "Project root: $PROJECT_ROOT"

# Build the karl binary
echo "Building karl..."
cd "$PROJECT_ROOT"
go build -o karl .

# Path to karl binary
KARL_BIN="$PROJECT_ROOT/karl"

if [ ! -f "$KARL_BIN" ]; then
    echo "Error: karl binary not found at $KARL_BIN"
    exit 1
fi

echo "Karl binary built at: $KARL_BIN"

# Create kernel directory
echo "Creating kernel directory at $KERNEL_DIR..."
mkdir -p "$KERNEL_DIR"

# Create kernel.json
cat > "$KERNEL_DIR/kernel.json" <<EOF
{
 "argv": [
  "$KARL_BIN",
  "kernel",
  "{connection_file}"
 ],
 "display_name": "Karl",
 "language": "karl"
}
EOF

echo "Kernel spec installed successfully!"
echo "Run 'jupyter kernelspec list' to verify."
echo "You can now start Jupyter Notebook/Lab and select the 'Karl' kernel."
