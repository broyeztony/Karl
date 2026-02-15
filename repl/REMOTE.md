# Karl Remote REPL

Connect to a Karl REPL running on a remote server over TCP.

## Quick Start

### Start the Server

```bash
# Start server on default port (localhost:9000)
./karl repl-server

# Start server on custom address
./karl repl-server --addr=0.0.0.0:9000

# Start server on specific port
./karl repl-server --addr=:8080
```

### Connect with Client

```bash
# Connect to local server
./karl repl-client localhost:9000

# Connect to remote server
./karl repl-client 192.168.1.100:9000

# With rlwrap for history support
rlwrap ./karl repl-client localhost:9000
```

## Alternative Connection Methods

You can also connect using standard tools:

```bash
# Using netcat
nc localhost 9000

# Using netcat with rlwrap
rlwrap nc localhost 9000

# Using telnet
telnet localhost 9000
```

## Features

- **Multi-client support**: Multiple clients can connect simultaneously
- **Isolated sessions**: Each client gets their own environment
- **Full REPL features**: All REPL commands work (`:help`, `:examples`, etc.)
- **Network transparent**: Works over any TCP connection

## Use Cases

### Development Team Collaboration
Share a REPL session with team members for pair programming or debugging.

### Remote Development
Run Karl on a server and connect from your local machine.

### Teaching/Demos
Instructor runs the server, students connect to follow along.

### Container/Cloud Deployment
Run Karl REPL in a container and connect from outside.

## Security Note

⚠️ **Warning**: The REPL server has no authentication. Only run it on trusted networks or use SSH tunneling for remote access.

### SSH Tunnel Example

```bash
# On your local machine, create SSH tunnel
ssh -L 9000:localhost:9000 user@remote-server

# On remote server, start Karl REPL server
./karl repl-server

# On local machine, connect through tunnel
./karl repl-client localhost:9000
```

## Examples

### Example 1: Local Development

Terminal 1:
```bash
$ ./karl repl-server
Karl REPL Server listening on localhost:9000
Connect with: ./karl repl-client localhost:9000
Or use: rlwrap nc localhost 9000
```

Terminal 2:
```bash
$ ./karl repl-client localhost:9000
Connected to Karl REPL Server at localhost:9000
Press Ctrl+C to disconnect

╔═══════════════════════════════════════╗
║   Karl REPL - Remote Session         ║
╚═══════════════════════════════════════╝

Connected to Karl REPL Server
Type expressions and press Enter to evaluate.
Commands: :help, :quit, :env, :examples

karl> let x = 42
karl> x * 2
84
karl> :quit
Goodbye!
```

### Example 2: Team Collaboration

```bash
# Team lead starts server
$ ./karl repl-server --addr=0.0.0.0:9000

# Team members connect
$ ./karl repl-client team-server.local:9000
```

## Command Reference

### Server

```bash
karl repl-server [OPTIONS]

Options:
  --addr string   Address to listen on (default "localhost:9000")
  -h, --help      Show help message

Examples:
  karl repl-server
  karl repl-server --addr=0.0.0.0:9000
  karl repl-server --addr=:8080
```

### Client

```bash
karl repl-client <host:port>

Arguments:
  host:port   Server address to connect to

Options:
  -h, --help  Show help message

Examples:
  karl repl-client localhost:9000
  karl repl-client 192.168.1.100:9000
```

## Troubleshooting

### Connection Refused
- Check that the server is running
- Verify the address and port are correct
- Check firewall settings

### Port Already in Use
- Another process is using the port
- Use a different port with `--addr=:PORT`

### No Response from Server
- Check network connectivity
- Verify firewall allows the connection
- Try using `telnet` or `nc` to test connectivity
