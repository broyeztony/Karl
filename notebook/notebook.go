// Package notebook provides a notebook system for Karl code execution.
// A notebook consists of a sequence of cells that can be executed in order,
// maintaining environment state across cell evaluations.
package notebook

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
)

// CellType defines the type of a notebook cell.
type CellType string

const (
	CodeCell CellType = "code"
	MarkdownCell CellType = "markdown"
)

// Cell represents a single cell in a notebook.
type Cell struct {
	Type     CellType              `json:"type"`
	Title    string                `json:"title,omitempty"`
	Source   string                `json:"source"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CellOutput represents the output of a cell execution.
type CellOutput struct {
	CellIndex int                   `json:"cell_index"`
	Type      string                `json:"type"`
	Value     string                `json:"value"`
	Error     *ExecutionError       `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time             `json:"timestamp"`
}

// ExecutionError represents an error that occurred during cell execution.
type ExecutionError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Notebook represents a Karl notebook.
type Notebook struct {
	Title       string                `json:"title"`
	Description string                `json:"description,omitempty"`
	Version     string                `json:"version,omitempty"`
	Cells       []Cell                `json:"cells"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewNotebook creates a new notebook with the given title.
func NewNotebook(title string) *Notebook {
	return &Notebook{
		Title:    title,
		Version:  "1.0",
		Cells:    []Cell{},
		Metadata: make(map[string]interface{}),
	}
}

// AddCodeCell adds a code cell to the notebook.
func (n *Notebook) AddCodeCell(source string, title string) {
	n.Cells = append(n.Cells, Cell{
		Type:     CodeCell,
		Title:    title,
		Source:   source,
		Metadata: make(map[string]interface{}),
	})
}

// AddMarkdownCell adds a markdown cell to the notebook.
func (n *Notebook) AddMarkdownCell(source string) {
	n.Cells = append(n.Cells, Cell{
		Type:     MarkdownCell,
		Source:   source,
		Metadata: make(map[string]interface{}),
	})
}

// LoadNotebook loads a notebook from a JSON file.
func LoadNotebook(filename string) (*Notebook, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read notebook: %w", err)
	}

	var nb Notebook
	if err := json.Unmarshal(data, &nb); err != nil {
		return nil, fmt.Errorf("failed to parse notebook: %w", err)
	}

	return &nb, nil
}

// SaveNotebook saves a notebook to a JSON file.
func (n *Notebook) SaveNotebook(filename string) error {
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal notebook: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write notebook: %w", err)
	}

	return nil
}

// Runner executes notebook cells and maintains evaluation context.
type Runner struct {
	env        *interpreter.Environment
	eval       *interpreter.Evaluator
	outputs    []CellOutput
	lastOutput interpreter.Value
}

// NewRunner creates a new notebook runner.
func NewRunner() *Runner {
	env := interpreter.NewBaseEnvironment()
	eval := interpreter.NewEvaluatorWithSourceAndFilename("", "<notebook>")
	return &Runner{
		env:     env,
		eval:    eval,
		outputs: []CellOutput{},
	}
}

// ExecuteNotebook executes all code cells in a notebook and returns outputs.
func (r *Runner) ExecuteNotebook(notebook *Notebook) ([]CellOutput, error) {
	r.outputs = []CellOutput{}

	for i, cell := range notebook.Cells {
		if cell.Type != CodeCell {
			continue
		}

		output, err := r.executeCell(i, cell)
		r.outputs = append(r.outputs, output)

		if err != nil {
			return r.outputs, err
		}
	}

	return r.outputs, nil
}

// ExecuteCell executes a single cell and returns its output.
func (r *Runner) ExecuteCell(cellIndex int, cell Cell) (CellOutput, error) {
	return r.executeCell(cellIndex, cell)
}

func (r *Runner) executeCell(cellIndex int, cell Cell) (CellOutput, error) {
	output := CellOutput{
		CellIndex: cellIndex,
		Type:      "result",
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Parse the cell source
	l := lexer.New(cell.Source)
	p := parser.New(l)
	program := p.ParseProgram()

	if errs := p.Errors(); len(errs) > 0 {
		errMsg := fmt.Sprintf("Parse error: %s", errs[0])
		output.Error = &ExecutionError{
			Message: errMsg,
			Type:    "ParseError",
		}
		return output, nil
	}

	// Evaluate the program
	result, _, err := r.eval.Eval(program, r.env)
	if err != nil {
		errMsg := fmt.Sprintf("Evaluation error: %v", err)
		output.Error = &ExecutionError{
			Message: errMsg,
			Type:    "RuntimeError",
		}
		return output, nil
	}

	// Store output if not nil/unit
	if result != nil {
		r.lastOutput = result
		if _, ok := result.(*interpreter.Unit); !ok {
			output.Value = result.Inspect()
		}
	}

	return output, nil
}

// GetOutputs returns all cell outputs from the last execution.
func (r *Runner) GetOutputs() []CellOutput {
	return r.outputs
}

// GetLastOutput returns the last output value.
func (r *Runner) GetLastOutput() interpreter.Value {
	return r.lastOutput
}

// GetEnvironment returns the current evaluation environment.
func (r *Runner) GetEnvironment() *interpreter.Environment {
	return r.env
}
