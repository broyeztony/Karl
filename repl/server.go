package repl

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"golang.org/x/term"
)

// Server starts a REPL server on the specified address (e.g., "localhost:9000")
func Server(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Karl REPL Server listening on %s\n", addr)
	fmt.Printf("Connect with: ./karl loom connect %s\n", addr)
	fmt.Printf("Or use: rlwrap nc %s\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection: %v\n", err)
			continue
		}

		// Handle each connection in a goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("New connection from %s\n", remoteAddr)

	// Send welcome message
	fmt.Fprintf(conn, "╔═══════════════════════════════════════╗\n")
	fmt.Fprintf(conn, "║   Karl REPL - Remote Session          ║\n")
	fmt.Fprintf(conn, "╚═══════════════════════════════════════╝\n")
	fmt.Fprintf(conn, "\n")
	fmt.Fprintf(conn, "Connected to Karl REPL Server\n")
	fmt.Fprintf(conn, "Type expressions and press Enter to evaluate.\n")
	fmt.Fprintf(conn, "Commands: :help, :quit, :clear, :env, :examples\n\n")

	// Start REPL session for this connection
	start(conn, conn, startOptions{showIntro: false})

	fmt.Printf("Connection closed from %s\n", remoteAddr)
}

// Client connects to a remote REPL server
func Client(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", addr, err)
	}
	defer conn.Close()

	fmt.Printf("Connected to Karl REPL Server at %s\n", addr)
	fmt.Printf("Press Ctrl+C to disconnect\n\n")

	restore, rawEnabled := enableClientRawMode(os.Stdin, os.Stdout)
	if rawEnabled {
		defer func() {
			if restoreErr := restore(); restoreErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to restore terminal mode: %v\n", restoreErr)
			}
		}()
	}

	serverOut := io.Writer(os.Stdout)
	if rawEnabled {
		serverOut = newTTYLineWriter(os.Stdout)
	}

	// Start two goroutines: one for reading from server, one for writing to server
	done := make(chan error, 2)

	// Read from server and write to stdout
	go func() {
		_, copyErr := io.Copy(serverOut, conn)
		done <- copyErr
	}()

	// Read from stdin and write to server
	go func() {
		_, copyErr := io.Copy(conn, os.Stdin)
		done <- copyErr
	}()

	// Wait for either goroutine to finish
	if copyErr := <-done; copyErr != nil && !errors.Is(copyErr, io.EOF) && !errors.Is(copyErr, net.ErrClosed) {
		return fmt.Errorf("repl stream copy failed: %w", copyErr)
	}

	return nil
}

func enableClientRawMode(stdin *os.File, stdout *os.File) (func() error, bool) {
	if stdin == nil || stdout == nil {
		return nil, false
	}
	if !term.IsTerminal(int(stdin.Fd())) || !term.IsTerminal(int(stdout.Fd())) {
		return nil, false
	}
	state, err := term.MakeRaw(int(stdin.Fd()))
	if err != nil {
		return nil, false
	}
	return func() error {
		return term.Restore(int(stdin.Fd()), state)
	}, true
}
