package spreadsheet

import (
	"fmt"
	"karl/interpreter"
	"karl/lexer"
	"karl/parser"
	"regexp"
	"strconv"
	"strings"
)

// SetCell updates a cell's raw value and triggers re-evaluation of the dependency graph.
func (s *Sheet) SetCell(id CellID, rawValue string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Pre-process for ranges: Convert "A1:A3" -> "[A1, A2, A3]"
	execValue := rawValue
	if strings.HasPrefix(rawValue, "=") {
		re := regexp.MustCompile(`([A-Z]+[0-9]+):([A-Z]+[0-9]+)`)
		execValue = re.ReplaceAllStringFunc(rawValue, func(match string) string {
			parts := strings.Split(match, ":")
			if len(parts) != 2 {
				return match
			}
			cells := expandRange(parts[0], parts[1])
			if len(cells) == 0 {
				return match
			}
			return fmt.Sprintf("[%s]", strings.Join(cells, ", "))
		})
	}

	cell, ok := s.Cells[id]
	if !ok {
		cell = &Cell{
			ID:           id,
			Dependencies: []CellID{},
			Dependents:   []CellID{},
		}
		s.Cells[id] = cell
	}
	
	cell.RawValue = rawValue // Store ORIGINAL value for display
	// But we need to execute the PROCESSED value.
	// We need a field for ExecValue, or we pass it to evaluateCell?
	// evaluateCell currently uses cell.RawValue. 
	// Let's add an internal ExecValue field to Cell struct or modify evaluateCell to accept code?
	// Modifying Cell struct requires touching sheet.go too. 
	// Simplest for now: Pass execValue to evaluateCell? 
	// But propagateUpdates calls evaluateCell(depCell) which re-reads RawValue.
	// So depCell needs to know its ExecValue.
	// Thus, we MUST store ExecValue in Cell.
	cell.ExecValue = execValue
	
	var deps []CellID
	if strings.HasPrefix(execValue, "=") {
		deps = extractDependencies(execValue)
	}
	
	s.updateDependencies(cell, deps)
	s.evaluateCell(cell)
	s.propagateUpdates(cell, make(map[CellID]bool))

	return nil
}

// expandRange converts "A1:A3" to []string{"A1", "A2", "A3"}
// Supports vertical (A1:A3), horizontal (A1:C1), and rectangular (A1:B2)
func expandRange(start, end string) []string {
	c1, r1, err1 := parseCellID(start)
	c2, r2, err2 := parseCellID(end)
	
	if err1 != nil || err2 != nil {
		return nil
	}
	
	// Normalize so c1<=c2 and r1<=r2
	if c1 > c2 { c1, c2 = c2, c1 }
	if r1 > r2 { r1, r2 = r2, r1 }
	
	var cells []string
	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			colStr := string(rune('A' + c - 1))
			cells = append(cells, fmt.Sprintf("%s%d", colStr, r))
		}
	}
	return cells
}

// parseCellID parses "A1" into col(1-based), row(1-based)
// Simple impl assuming single letter column 'A'-'Z' for now usually, 
// but let's robustly handle "AA".
func parseCellID(id string) (int, int, error) {
	// Split letters and numbers
	re := regexp.MustCompile(`^([A-Z]+)([0-9]+)$`)
	parts := re.FindStringSubmatch(id)
	if len(parts) != 3 {
		return 0, 0, fmt.Errorf("invalid cell id")
	}
	colStr := parts[1]
	rowStr := parts[2]
	
	row, err := strconv.Atoi(rowStr)
	if err != nil {
		return 0, 0, err
	}
	
	col := 0
	for _, ch := range colStr {
		col = col*26 + int(ch-'A'+1)
	}
	
	return col, row, nil
}

// updateDependencies updates the graph edges.
func (s *Sheet) updateDependencies(cell *Cell, newDeps []CellID) {
    // Remove this cell from old dependencies' dependents list
    for _, oldDepID := range cell.Dependencies {
        if oldDep, ok := s.Cells[oldDepID]; ok {
            oldDep.removeDependent(cell.ID)
        }
    }
    
    // Set new dependencies
    cell.Dependencies = newDeps
    
    // Add this cell to new dependencies' dependents list
    for _, newDepID := range newDeps {
        depCell, ok := s.Cells[newDepID]
        if !ok {
             // Create empty cell if referenced but not exists
             depCell = &Cell{
                 ID: newDepID,
                 Dependencies: []CellID{},
                 Dependents:   []CellID{},
             }
             s.Cells[newDepID] = depCell
        }
        depCell.addDependent(cell.ID)
    }
}

func (c *Cell) addDependent(id CellID) {
    for _, dep := range c.Dependents {
        if dep == id {
            return
        }
    }
    c.Dependents = append(c.Dependents, id)
}

func (c *Cell) removeDependent(id CellID) {
    newDeps := []CellID{}
    for _, dep := range c.Dependents {
        if dep != id {
            newDeps = append(newDeps, dep)
        }
    }
    c.Dependents = newDeps
}

// evaluateCell executes the cell's raw value using the Karl interpreter.
// Caller must hold s.mu.
func (s *Sheet) evaluateCell(cell *Cell) {
    // Update environment with dependency values
    for _, depID := range cell.Dependencies {
        depCell := s.Cells[depID]
        if depCell.Value != nil {
            // Check if Value is a valid interpreter.Value
            if val, ok := depCell.Value.(interpreter.Value); ok {
                 s.env.Define(string(depCell.ID), val)
            }
        }
    }

	// Use ExecValue (pre-processed) for evaluation
	execValue := cell.ExecValue
	if execValue == "" {
		execValue = cell.RawValue // Fallback
	}

	var val interpreter.Value
	var err error

	if strings.HasPrefix(execValue, "=") {
		code := execValue[1:]
		// Evaluate as code
		l := lexer.New(code)
		p := parser.New(l)
		program := p.ParseProgram()
		
		if len(p.Errors()) > 0 {
			cell.Error = fmt.Errorf("Parse error: %v", p.Errors())
			cell.Value = nil
			return // Should return? Or set error and continue?
		}
		// Evaluate
		val, _, err = s.eval.Eval(program, s.env)
		if err != nil {
			fmt.Printf("Error evaluating cell %s: %v\nCode: %s\n", cell.ID, err, execValue)
			cell.Error = err
			cell.Value = nil
			return
		}
		
		// Store result
		cell.Value = val
		cell.Error = nil
	} else {
		// Try Number
		if f, err := strconv.ParseFloat(execValue, 64); err == nil {
			// Check if integer
			if !strings.Contains(execValue, ".") && !strings.Contains(strings.ToLower(execValue), "e") {
				if i, err := strconv.ParseInt(execValue, 10, 64); err == nil {
					cell.Value = &interpreter.Integer{Value: i}
				} else {
					cell.Value = &interpreter.Float{Value: f}
				}
			} else {
				cell.Value = &interpreter.Float{Value: f}
			}
		} else if execValue == "true" {
			cell.Value = &interpreter.Boolean{Value: true}
		} else if execValue == "false" {
			cell.Value = &interpreter.Boolean{Value: false}
		} else {
			// String or Empty
			if execValue == "" {
				cell.Value = nil // Empty
			} else {
				cell.Value = &interpreter.String{Value: execValue}
			}
		}
		cell.Error = nil
		if cell.Value != nil {
			val = cell.Value.(interpreter.Value)
		}
	}
    
    // Update the environment with this cell's new value
	if val != nil {
		s.env.Define(string(cell.ID), val)
	}
}

func (s *Sheet) propagateUpdates(cell *Cell, visited map[CellID]bool) {
    if visited[cell.ID] {
        return // Cycle detected or already visited
    }
    visited[cell.ID] = true
    
    for _, depID := range cell.Dependents {
        if depCell, ok := s.Cells[depID]; ok {
            s.evaluateCell(depCell)
            // Recursively propagate
            s.propagateUpdates(depCell, visited)
        }
    }
}

var cellIDRegex = regexp.MustCompile(`[A-Z]+[0-9]+`)

func extractDependencies(code string) []CellID {
    matches := cellIDRegex.FindAllString(code, -1)
    unique := make(map[CellID]bool)
    var deps []CellID
    for _, m := range matches {
        id := CellID(m)
        if !unique[id] {
            unique[id] = true
            deps = append(deps, id)
        }
    }
    return deps
}
