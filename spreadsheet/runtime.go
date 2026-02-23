package spreadsheet

import "sort"

// Runtime is an in-process spreadsheet runtime used by local integrations
// (for example the browser WASM worker).
type Runtime struct {
	server *Server
}

func NewRuntime() *Runtime {
	return &Runtime{
		server: NewServer(),
	}
}

func (r *Runtime) Clear() {
	r.server.Sheet.Clear()
}

func (r *Runtime) LoadExample(name string) {
	switch name {
	case "heavy":
		r.server.populateHeavy()
	case "syntax":
		r.server.populateSyntax()
	case "matrix":
		r.server.populateMatrix()
	case "ranges":
		r.server.populateRanges()
	case "factorial":
		r.server.populateFactorial()
	default:
		r.server.populateIntro()
	}
}

func (r *Runtime) Snapshot() []UpdateResponse {
	r.server.Sheet.mu.RLock()
	ids := make([]string, 0, len(r.server.Sheet.Cells))
	for id := range r.server.Sheet.Cells {
		ids = append(ids, string(id))
	}
	r.server.Sheet.mu.RUnlock()

	sort.Strings(ids)
	out := make([]UpdateResponse, 0, len(ids))
	for _, rawID := range ids {
		id := CellID(rawID)
		r.server.Sheet.mu.RLock()
		cell, ok := r.server.Sheet.Cells[id]
		r.server.Sheet.mu.RUnlock()
		if !ok {
			continue
		}
		out = append(out, r.server.createUpdateResponse(cell))
	}
	return out
}

func (r *Runtime) UpdateCell(id string, value string) ([]UpdateResponse, error) {
	cellID := CellID(id)
	if err := r.server.Sheet.SetCell(cellID, value); err != nil {
		return nil, err
	}

	affected := make(map[CellID]bool)
	r.server.collectAffected(cellID, affected)

	ordered := make([]string, 0, len(affected))
	for affectedID := range affected {
		ordered = append(ordered, string(affectedID))
	}
	sort.Strings(ordered)

	out := make([]UpdateResponse, 0, len(ordered))
	for _, rawID := range ordered {
		id := CellID(rawID)
		r.server.Sheet.mu.RLock()
		cell, ok := r.server.Sheet.Cells[id]
		r.server.Sheet.mu.RUnlock()
		if !ok {
			continue
		}
		out = append(out, r.server.createUpdateResponse(cell))
	}
	return out, nil
}
