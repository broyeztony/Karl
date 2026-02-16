package kernel

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-zeromq/zmq4"
	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

// ConnectionInfo holds the connection file configuration
type ConnectionInfo struct {
	SignatureScheme string `json:"signature_scheme"`
	Transport       string `json:"transport"`
	StdinPort       int    `json:"stdin_port"`
	ControlPort     int    `json:"control_port"`
	IOPubPort       int    `json:"iopub_port"`
	HBPort          int    `json:"hb_port"`
	ShellPort       int    `json:"shell_port"`
	Key             string `json:"key"`
	IP              string `json:"ip"`
}

// Header represents the Jupyter message header
type Header struct {
	MsgID    string `json:"msg_id"`
	Username string `json:"username"`
	Session  string `json:"session"`
	Date     string `json:"date"`
	MsgType  string `json:"msg_type"`
	Version  string `json:"version"`
}

// Message represents a Jupyter protocol message
type Message struct {
	Header       Header                 `json:"header"`
	ParentHeader Header                 `json:"parent_header"`
	Metadata     map[string]interface{} `json:"metadata"`
	Content      map[string]interface{} `json:"content"`
}

// Kernel represents the running Jupyter kernel
type Kernel struct {
	config     ConnectionInfo
	hb         zmq4.Socket
	shell      zmq4.Socket
	control    zmq4.Socket
	iopub      zmq4.Socket
	stdin      zmq4.Socket
	sockets    []zmq4.Socket
	shutdown   chan struct{}
	eval       *interpreter.Evaluator
	env        *interpreter.Environment
	executionCount int
	mu         sync.Mutex
}

// NewKernel creates a new Jupyter kernel instance
func NewKernel(configPath string) (*Kernel, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read connection file: %w", err)
	}

	var config ConnectionInfo
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse connection file: %w", err)
	}

	k := &Kernel{
		config:   config,
		shutdown: make(chan struct{}),
	}

	// Initialize Karl interpreter
	k.env = interpreter.NewBaseEnvironment()
	k.eval = interpreter.NewEvaluatorWithSourceAndFilename("", "<jupyter>")

	return k, nil
}

// Start starts the kernel and its ZeroMQ sockets
func (k *Kernel) Start() error {
	var err error
	ctx := context.Background()

	// Helper to create sockets
	// Helper to create sockets
	createSocket := func(sockType zmq4.SocketType, port int) (zmq4.Socket, error) {
		// Create socket based on type
		var sock zmq4.Socket
		switch sockType {
		case zmq4.Rep:
			sock = zmq4.NewRep(ctx)
		case zmq4.Router:
			sock = zmq4.NewRouter(ctx)
		case zmq4.Pub:
			sock = zmq4.NewPub(ctx)
		default:
			return nil, fmt.Errorf("unsupported socket type: %v", sockType)
		}
		
		addr := fmt.Sprintf("%s://%s:%d", k.config.Transport, k.config.IP, port)
		if err := sock.Listen(addr); err != nil {
			return nil, fmt.Errorf("failed to bind to %s: %w", addr, err)
		}
		return sock, nil
	}

	// Heartbeat
	k.hb, err = createSocket(zmq4.Rep, k.config.HBPort)
	if err != nil {
		return err
	}
	go k.handleHeartbeat()

	// Shell
	k.shell, err = createSocket(zmq4.Router, k.config.ShellPort)
	if err != nil {
		return err
	}

	// IOPub
	k.iopub, err = createSocket(zmq4.Pub, k.config.IOPubPort)
	if err != nil {
		return err
	}

	// Control
	k.control, err = createSocket(zmq4.Router, k.config.ControlPort)
	if err != nil {
		return err
	}

	// Stdin
	k.stdin, err = createSocket(zmq4.Router, k.config.StdinPort)
	if err != nil {
		return err
	}

	k.sockets = []zmq4.Socket{k.hb, k.shell, k.control, k.iopub, k.stdin}

	log.Printf("Kernel started, listening on ports: HB=%d Shell=%d IOPub=%d Control=%d Stdin=%d",
		k.config.HBPort, k.config.ShellPort, k.config.IOPubPort, k.config.ControlPort, k.config.StdinPort)

	// Main loop for shell and control channels
	go k.handleShell()
	go k.handleControl()

	<-k.shutdown
	return nil
}

// Stop stops the kernel
func (k *Kernel) Stop() {
	close(k.shutdown)
	for _, sock := range k.sockets {
		sock.Close()
	}
}

func (k *Kernel) handleHeartbeat() {
	for {
		msg, err := k.hb.Recv()
		if err != nil {
			return
		}
		k.hb.Send(msg)
	}
}

func (k *Kernel) handleShell() {
	for {
		identities, msg, err := k.receiveMessage(k.shell)
		if err != nil {
			log.Printf("Error receiving shell message: %v", err)
			continue
		}
		
		switch msg.Header.MsgType {
		case "kernel_info_request":
			k.handleKernelInfoRequest(msg, identities)
		case "execute_request":
			k.handleExecuteRequest(msg, identities)
		case "shutdown_request":
			k.handleShutdownRequest(msg, identities)
		default:
			log.Printf("Unknown shell message type: %s", msg.Header.MsgType)
		}
	}
}

func (k *Kernel) handleControl() {
	for {
		identities, msg, err := k.receiveMessage(k.control)
		if err != nil {
			log.Printf("Error receiving control message: %v", err)
			continue
		}

		switch msg.Header.MsgType {
		case "shutdown_request":
			k.handleShutdownRequest(msg, identities)
		default:
			log.Printf("Unknown control message type: %s", msg.Header.MsgType)
		}
	}
}

// receiveMessage reads a full Jupyter message from a socket
func (k *Kernel) receiveMessage(sock zmq4.Socket) ([][]byte, *Message, error) {
	// Jupyter message structure:
	// [identities...] <DELIMITER> <HMAC> <Header> <ParentHeader> <Metadata> <Content>
	
	msg, err := sock.Recv()
	if err != nil {
		return nil, nil, err
	}
	
	frames := msg.Frames
	delimiterParams := -1
	for i, frame := range frames {
		if string(frame) == "<IDS|MSG>" {
			delimiterParams = i
			break
		}
	}

	if delimiterParams == -1 {
		return nil, nil, fmt.Errorf("message delimiter not found")
	}

	identities := frames[:delimiterParams]
	// signature := string(frames[delimiterParams+1])
	headerBytes := frames[delimiterParams+2]
	parentHeaderBytes := frames[delimiterParams+3]
	metadataBytes := frames[delimiterParams+4]
	contentBytes := frames[delimiterParams+5]

	var m Message
	if err := json.Unmarshal(headerBytes, &m.Header); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(parentHeaderBytes, &m.ParentHeader); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(metadataBytes, &m.Metadata); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(contentBytes, &m.Content); err != nil {
		return nil, nil, err
	}

	return identities, &m, nil
}

func (k *Kernel) sendMessage(sock zmq4.Socket, msg *Message, identities ...[]byte) error {
	header, _ := json.Marshal(msg.Header)
	parentHeader, _ := json.Marshal(msg.ParentHeader)
	metadata, _ := json.Marshal(msg.Metadata)
	content, _ := json.Marshal(msg.Content)

	// Calculate HMAC signature
	mac := hmac.New(sha256.New, []byte(k.config.Key))
	mac.Write(header)
	mac.Write(parentHeader)
	mac.Write(metadata)
	mac.Write(content)
	signature := hex.EncodeToString(mac.Sum(nil))

	frames := [][]byte{
		[]byte("<IDS|MSG>"),
		[]byte(signature),
		header,
		parentHeader,
		metadata,
		content,
	}
	
	// Prepend identities if provided (needed for Router sockets)
	allFrames := append(identities, frames...)
	
	zmsg := zmq4.NewMsgFrom(allFrames...)
	return sock.Send(zmsg)
}

func (k *Kernel) handleKernelInfoRequest(msg *Message, identities [][]byte) {
	content := map[string]interface{}{
		"protocol_version":       "5.3",
		"implementation":         "karl-kernel",
		"implementation_version": "0.1.0",
		"language_info": map[string]interface{}{
			"name":           "karl",
			"version":        "0.1.0",
			"mimetype":       "text/x-karl",
			"file_extension": ".k",
		},
		"banner": "Karl Programming Language Kernel",
	}

	reply := &Message{
		Header: Header{
			MsgID:    newUUID(),
			Username: "kernel",
			Session:  msg.Header.Session,
			MsgType:  "kernel_info_reply",
			Version:  "5.3",
			Date:     time.Now().Format(time.RFC3339),
		},
		ParentHeader: msg.Header,
		Metadata:     make(map[string]interface{}),
		Content:      content,
	}

	k.sendMessage(k.shell, reply, identities...) 
}

func (k *Kernel) handleShutdownRequest(msg *Message, identities [][]byte) {
	restart := msg.Content["restart"].(bool)
	
	reply := &Message{
		Header: Header{
			MsgID:    newUUID(),
			Username: "kernel",
			Session:  msg.Header.Session,
			MsgType:  "shutdown_reply",
			Version:  "5.3",
			Date:     time.Now().Format(time.RFC3339),
		},
		ParentHeader: msg.Header,
		Content: map[string]interface{}{
			"restart": restart,
		},
	}
	
	k.sendMessage(k.shell, reply, identities...)
	if !restart {
		k.Stop()
	}
}

func (k *Kernel) handleExecuteRequest(msg *Message, identities [][]byte) {
	code := msg.Content["code"].(string)
	k.mu.Lock()
	k.executionCount++
	execCount := k.executionCount
	k.mu.Unlock()

	// Publish execute_input status
	k.publishStatus("busy", msg.Header)
	
	k.publishExecuteInput(code, execCount, msg.Header)

	// Execute Code
	var result string
	var errResult error
	
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		errResult = fmt.Errorf("Parse Error: %v", p.Errors())
	} else {
		val, _, err := k.eval.Eval(program, k.env)
		if err != nil {
			errResult = err
		} else if val != nil {
			if _, ok := val.(*interpreter.Unit); !ok {
				result = val.Inspect()
			}
		}
	}

	if errResult != nil {
		// Publish Error
		errorContent := map[string]interface{}{
			"ename":    "Error",
			"evalue":   errResult.Error(),
			"traceback": []string{errResult.Error()},
		}
		
		errorMsg := &Message{
			Header: Header{
				MsgID:    newUUID(),
				Username: "kernel",
				Session:  msg.Header.Session,
				MsgType:  "error",
				Version:  "5.3",
				Date:     time.Now().Format(time.RFC3339),
			},
			ParentHeader: msg.Header,
			Content:      errorContent,
		}
		k.sendMessage(k.iopub, errorMsg)

		// Send execute_reply (error)
		reply := &Message{
			Header: Header{
				MsgID:    newUUID(),
				Username: "kernel",
				Session:  msg.Header.Session,
				MsgType:  "execute_reply",
				Version:  "5.3",
				Date:     time.Now().Format(time.RFC3339),
			},
			ParentHeader: msg.Header,
			Content: map[string]interface{}{
				"status":           "error",
				"execution_count":  execCount,
				"ename":            "Error",
				"evalue":           errResult.Error(),
				"traceback":        []string{errResult.Error()},
			},
		}
		k.sendMessage(k.shell, reply, identities...)

	} else {
		// Publish Result (execute_result) if there is output
		if result != "" {
			resultContent := map[string]interface{}{
				"execution_count": execCount,
				"data": map[string]interface{}{
					"text/plain": result,
				},
				"metadata": map[string]interface{}{},
			}
			
			resultMsg := &Message{
				Header: Header{
					MsgID:    newUUID(),
					Username: "kernel",
					Session:  msg.Header.Session,
					MsgType:  "execute_result",
					Version:  "5.3",
					Date:     time.Now().Format(time.RFC3339),
				},
				ParentHeader: msg.Header,
				Content:      resultContent,
			}
			k.sendMessage(k.iopub, resultMsg)
		}

		// Send execute_reply (ok)
		reply := &Message{
			Header: Header{
				MsgID:    newUUID(),
				Username: "kernel",
				Session:  msg.Header.Session,
				MsgType:  "execute_reply",
				Version:  "5.3",
				Date:     time.Now().Format(time.RFC3339),
			},
			ParentHeader: msg.Header,
			Content: map[string]interface{}{
				"status":           "ok",
				"execution_count":  execCount,
				"payload":          []interface{}{},
				"user_expressions": map[string]interface{}{},
			},
		}
		k.sendMessage(k.shell, reply, identities...)
	}

	k.publishStatus("idle", msg.Header)
}

func (k *Kernel) publishStatus(status string, parentHeader Header) {
	content := map[string]interface{}{
		"execution_state": status,
	}
	msg := &Message{
		Header: Header{
			MsgID:    newUUID(),
			Username: "kernel",
			Session:  parentHeader.Session,
			MsgType:  "status",
			Version:  "5.3",
			Date:     time.Now().Format(time.RFC3339),
		},
		ParentHeader: parentHeader,
		Content:      content,
	}
	k.sendMessage(k.iopub, msg)
}

func (k *Kernel) publishExecuteInput(code string, count int, parentHeader Header) {
	content := map[string]interface{}{
		"code":            code,
		"execution_count": count,
	}
	msg := &Message{
		Header: Header{
			MsgID:    newUUID(),
			Username: "kernel",
			Session:  parentHeader.Session,
			MsgType:  "execute_input",
			Version:  "5.3",
			Date:     time.Now().Format(time.RFC3339),
		},
		ParentHeader: parentHeader,
		Content:      content,
	}
	k.sendMessage(k.iopub, msg)
}

func newUUID() string {
	// Simple UUID generation for demo
	b := make([]byte, 16)
	// rand.Read(b) - using time-based for now to avoid imports
	t := time.Now().UnixNano()
	return fmt.Sprintf("%x-%x", t, b)
}
