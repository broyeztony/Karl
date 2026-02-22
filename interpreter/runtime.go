package interpreter

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
)

// runtimeState is shared across Evaluator instances (main file, imported modules,
// spawned tasks). It lets us inspect task outcomes at the end of a run.
type runtimeState struct {
	mu                sync.Mutex
	tasks             map[*Task]struct{}
	taskFailurePolicy string
	fatalTaskFailure  error
	fatalOnce         sync.Once
	fatalCh           chan struct{}
	argv              []string
	programPath       *string
	environ           []string
	envMap            map[string]string
	input             io.Reader
	inputReader       *bufio.Reader
	inputMu           sync.Mutex
	sqlDriver         string
}

func newRuntimeState() *runtimeState {
	envSnapshot := os.Environ()
	return &runtimeState{
		tasks:             make(map[*Task]struct{}),
		taskFailurePolicy: TaskFailurePolicyFailFast,
		fatalCh:           make(chan struct{}),
		argv:              []string{},
		environ:           cloneStrings(envSnapshot),
		envMap:            makeEnvMap(envSnapshot),
		input:             os.Stdin,
		sqlDriver:         "pgx",
	}
}

func (r *runtimeState) registerTask(t *Task) {
	if r == nil || t == nil {
		return
	}
	r.mu.Lock()
	r.tasks[t] = struct{}{}
	r.mu.Unlock()
}

func (r *runtimeState) snapshotTasks() []*Task {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	out := make([]*Task, 0, len(r.tasks))
	for t := range r.tasks {
		out = append(out, t)
	}
	r.mu.Unlock()
	return out
}

func (r *runtimeState) setProgramArgs(args []string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.argv = cloneStrings(args)
	r.mu.Unlock()
}

func (r *runtimeState) snapshotProgramArgs() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	out := cloneStrings(r.argv)
	r.mu.Unlock()
	return out
}

func (r *runtimeState) setProgramPath(path string) {
	if r == nil {
		return
	}
	p := path
	r.mu.Lock()
	r.programPath = &p
	r.mu.Unlock()
}

func (r *runtimeState) clearProgramPath() {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.programPath = nil
	r.mu.Unlock()
}

func (r *runtimeState) getProgramPath() (string, bool) {
	if r == nil {
		return "", false
	}
	r.mu.Lock()
	path := r.programPath
	r.mu.Unlock()
	if path == nil {
		return "", false
	}
	return *path, true
}

func (r *runtimeState) setEnviron(environ []string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.environ = cloneStrings(environ)
	r.envMap = makeEnvMap(environ)
	r.mu.Unlock()
}

func (r *runtimeState) snapshotEnviron() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	out := cloneStrings(r.environ)
	r.mu.Unlock()
	return out
}

func (r *runtimeState) lookupEnv(name string) (string, bool) {
	if r == nil {
		return "", false
	}
	r.mu.Lock()
	value, ok := r.envMap[name]
	r.mu.Unlock()
	return value, ok
}

func (r *runtimeState) setInput(input io.Reader) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.input = input
	r.inputReader = nil
	r.mu.Unlock()
}

func (r *runtimeState) readLine() (string, bool, error) {
	if r == nil {
		return "", false, nil
	}

	r.inputMu.Lock()
	defer r.inputMu.Unlock()

	reader, err := r.inputBufReader()
	if err != nil {
		return "", false, err
	}

	line, err := reader.ReadString('\n')
	switch {
	case err == nil:
		return trimLineEnding(line), true, nil
	case err == io.EOF && line != "":
		return trimLineEnding(line), true, nil
	case err == io.EOF:
		return "", false, nil
	default:
		return "", false, err
	}
}

func (r *runtimeState) setSQLDriver(driver string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.sqlDriver = driver
	r.mu.Unlock()
}

func (r *runtimeState) getSQLDriver() string {
	if r == nil {
		return "pgx"
	}
	r.mu.Lock()
	driver := r.sqlDriver
	r.mu.Unlock()
	if driver == "" {
		return "pgx"
	}
	return driver
}

func (r *runtimeState) inputBufReader() (*bufio.Reader, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.inputReader != nil {
		return r.inputReader, nil
	}
	if r.input == nil {
		return nil, &RuntimeError{Message: "stdin unavailable"}
	}
	r.inputReader = bufio.NewReader(r.input)
	return r.inputReader, nil
}

func cloneStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func makeEnvMap(environ []string) map[string]string {
	out := make(map[string]string, len(environ))
	for _, entry := range environ {
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		out[key] = value
	}
	return out
}

func trimLineEnding(line string) string {
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line
}
