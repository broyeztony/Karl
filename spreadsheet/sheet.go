package spreadsheet

import (
	"karl/interpreter"
	"sync"
)

// CellID represents a cell identifier like "A1", "B2".
type CellID string

// Cell represents a single cell in the spreadsheet.
type Cell struct {
	ID        CellID
	RawValue  string      // The user input (e.g. "=A1:A3")
	ExecValue string      // The executed code (e.g. "=[A1, A2, A3]")
	Value     interface{} // The evaluated result
	Error     error       // Any error during evaluation
	
	// Dependency tracking
	Dependencies []CellID // Cells that this cell depends on
	Dependents   []CellID // Cells that depend on this cell
}

// Sheet represents a collection of cells and their relationships.
type Sheet struct {
	Cells map[CellID]*Cell
	mu    sync.RWMutex
	
	// Interpreter context for this sheet
	env  *interpreter.Environment
	eval *interpreter.Evaluator
}

// NewSheet creates a new empty spreadsheet.
func NewSheet() *Sheet {
	s := &Sheet{
		Cells: make(map[CellID]*Cell),
		env:   interpreter.NewBaseEnvironment(),
		// We'll initialize the evaluator later or per-execution if needed,
		// but typically we want one persistent context.
		eval: interpreter.NewEvaluatorWithSourceAndFilename("", "<spreadsheet>"),
	}
	return s
}

// Clear removes all cells from the sheet.
func (s *Sheet) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Cells = make(map[CellID]*Cell)
	// We might want to keep environment? Or clear it? 
	// Probably clear environment too to remove defined variables.
	s.env = interpreter.NewBaseEnvironment()
}

// GetCell returns a cell by ID, creating it if it doesn't exist.
func (s *Sheet) GetCell(id CellID) *Cell {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if cell, ok := s.Cells[id]; ok {
		return cell
	}
	
	cell := &Cell{
		ID:           id,
		Dependencies: []CellID{},
		Dependents:   []CellID{},
	}
	s.Cells[id] = cell
	return cell
}



// Helper to convert "A1" to (0, 0) coordinates if needed, 
// though we might stick to string IDs for simplicity in the map.
