# Karl Jupyter Kernel

This directory contains the implementation of a Jupyter kernel for the Karl programming language.

## Installation

To install the kernel, run the provided installation script:

```bash
./kernel/install.sh
```

This will:
1. Build the `karl` binary.
2. Create the kernel directory at `~/Library/Jupyter/kernels/karl` (on macOS).
3. Install the `kernel.json` spec file pointing to your `karl` binary.

## Usage

Once installed, you can start Jupyter Notebook or Jupyter Lab and select "Karl" from the kernel list when creating a new notebook.

```bash
jupyter notebook
# or
jupyter lab
```

## Features

- **Code Execution**: Run Karl code cells and see the output.
- **State Persistence**: Variables and functions defined in one cell are available in subsequent cells.
- **Error Handling**: Syntax errors and runtime errors are reported in the notebook.
- **Output Types**: Supports text output (more types coming soon).

## Manual Testing

You can manually test the kernel startup using a connection file:

1. Create a `connection.json` file (example provided in `examples/connection.json`).
2. Run the kernel:
   ```bash
   ./karl kernel connection.json
   ```
3. The kernel should start and listen on the configured ports.

## Development

The kernel implementation relies on ZeroMQ for communication with the Jupyter frontend. It uses `github.com/go-zeromq/zmq4`.

Key files:
- `kernel/kernel.go`: Main kernel logic, message handling, and ZeroMQ socket management.
- `kernel/install.sh`: Installation script.
