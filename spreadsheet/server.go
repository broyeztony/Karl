package spreadsheet

import (
	"encoding/json"
	"fmt"
	"karl/interpreter"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for local dev
	},
}

type Server struct {
	Sheet   *Sheet
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewServer() *Server {
	s := &Server{
		Sheet:   NewSheet(),
		clients: make(map[*websocket.Conn]bool),
	}
	s.populateIntro()
	return s
}

func (s *Server) mustSetCell(id CellID, rawValue string) {
	if err := s.Sheet.SetCell(id, rawValue); err != nil {
		log.Printf("set cell %s failed: %v", id, err)
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	// Send initial state
	s.sendInitialState(conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var req UpdateRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Println("JSON error:", err)
			continue
		}

		if req.Type == "update_cell" {
			s.handleUpdate(req)
		} else if req.Type == "clear" {
			s.Sheet.Clear()
			s.broadcastAll()
		} else if req.Type == "load_intro" {
			s.populateIntro()
			s.broadcastAll()
		} else if req.Type == "load_example" {
			if req.Example == "heavy" {
				s.populateHeavy()
			} else if req.Example == "syntax" {
				s.populateSyntax()
			} else if req.Example == "matrix" {
				s.populateMatrix()
			} else if req.Example == "ranges" {
				s.populateRanges()
			} else if req.Example == "factorial" {
				s.populateFactorial()
			} else {
				s.populateIntro()
			}
			s.broadcastAll()
		}
	}
}

func (s *Server) broadcastAll() {
	s.Sheet.mu.RLock()
	defer s.Sheet.mu.RUnlock()

	// Better: Send "reset" message.
	resetMsg := UpdateResponse{
		Type: "reset",
	}

	s.mu.Lock()
	for client := range s.clients {
		if err := client.WriteJSON(resetMsg); err != nil {
			log.Printf("reset write failed: %v", err)
			_ = client.Close()
			delete(s.clients, client)
		}
	}
	s.mu.Unlock()

	// Send new state
	for _, cell := range s.Sheet.Cells {
		resp := s.createUpdateResponse(cell)
		s.mu.Lock()
		for client := range s.clients {
			if err := client.WriteJSON(resp); err != nil {
				log.Printf("broadcast write failed: %v", err)
				_ = client.Close()
				delete(s.clients, client)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Server) populateIntro() {
	s.Sheet.Clear()

	// Header
	s.mustSetCell("A1", "🚀 Karl Sheets")
	s.mustSetCell("B1", "Interactivty Demo")

	// 1. Math
	s.mustSetCell("A3", "1. Math")
	s.mustSetCell("B3", "10")
	s.mustSetCell("C3", "32")
	s.mustSetCell("D3", "= B3 + C3")
	s.mustSetCell("E3", "<- Sum")

	// 2. Logic
	s.mustSetCell("A5", "2. Logic")
	s.mustSetCell("B5", "true")
	s.mustSetCell("C5", "= if B5 { \"Yes\" } else { \"No\" }")
	s.mustSetCell("D5", "<- Change B5!")

	// 3. Functions
	s.mustSetCell("A7", "3. Functions")
	s.mustSetCell("B7", "= (x) -> x * x")
	s.mustSetCell("C7", "= B7(10)")
	s.mustSetCell("D7", "<- Square(10)")

	// 4. Comparison
	s.mustSetCell("A9", "4. Comparison")
	s.mustSetCell("B9", "= 10 > 5")
	s.mustSetCell("C9", "= 10 < 5")
	s.mustSetCell("D9", "<- True/False")

	s.mustSetCell("A11", "5. Lists")
	s.mustSetCell("B11", "= [1, 2, 3, 4]")
	s.mustSetCell("C11", "= B11.length")
	s.mustSetCell("D11", "= B11[0] + B11[3]")

	// 6. Chain
	s.mustSetCell("A13", "6. Chain")
	s.mustSetCell("B13", "1")
	s.mustSetCell("C13", "= B13 + 1")
	s.mustSetCell("D13", "= C13 * 2")
	s.mustSetCell("E13", "= D13 * 10")
}

func (s *Server) populateHeavy() {
	s.Sheet.Clear()

	s.mustSetCell("A1", "🚀 Heavy Computation")
	s.mustSetCell("B1", "1000 chained cells")

	// Chain 1000 calculations across 20 columns and 50 rows
	// A2 = 1
	// B2 = A2 + 1
	// ...

	val := 1
	s.mustSetCell("A2", fmt.Sprintf("%d", val))

	rows := 50
	cols := 20
	// We'll fill columns A-T (1-20)
	// Simply linear chain: B2 = A2 + 1, C2 = B2 + 1...
	// Then A3 = T2 + 1

	prev := "A2"
	for r := 2; r < rows+2; r++ {
		for c := 1; c <= cols; c++ {
			if r == 2 && c == 1 {
				continue // Initial value set
			}
			colName := string(rune('A' + c - 1))
			id := CellID(fmt.Sprintf("%s%d", colName, r))

			formula := fmt.Sprintf("= %s + 1", prev)
			s.mustSetCell(id, formula)
			prev = string(id)
		}
	}
}

func (s *Server) populateSyntax() {
	s.Sheet.Clear()

	s.mustSetCell("A1", "🧠 Karl Syntax Demo")

	// 1. First Class Functions
	s.mustSetCell("A3", "1. Functions")
	s.mustSetCell("B3", "= (x) -> x * 2")
	s.mustSetCell("C3", "= (f, x) -> f(x) + 1")
	s.mustSetCell("D3", "= C3(B3, 10)") // (10*2)+1 = 21
	s.mustSetCell("E3", "<- Higher Order")

	// 2. Maps & Objects
	s.mustSetCell("A5", "2. Objects")
	s.mustSetCell("B5", "= {name: \"Karl\", ver: 1.0}")
	s.mustSetCell("D5", "= B5[\"ver\"] ? null")
	s.mustSetCell("E5", "= keys(B5)") // keys return order implementation specific

	// 3. List Processing
	s.mustSetCell("A7", "3. List Ops")
	s.mustSetCell("B7", "= [1, 2, 3, 4, 5]")
	s.mustSetCell("C7", "= B7.map((x) -> x * x)")
	s.mustSetCell("D7", "= B7.filter((x) -> x > 2)")
	s.mustSetCell("E7", "= B7.reduce((acc, x) -> acc + x, 0)")

	// 4. String Manipulation
	s.mustSetCell("A9", "4. Strings")
	s.mustSetCell("B9", "Hello World")
	s.mustSetCell("C9", "= toUpper(B9)")
	s.mustSetCell("D9", "= split(B9, \" \")")
	s.mustSetCell("E9", "= replace(B9, \"l\", \"1\")")

	// 5. Conditionals & Pattern Matching (simulated)
	s.mustSetCell("A11", "5. Logic")
	s.mustSetCell("B11", "admin")
	s.mustSetCell("C11", "= if B11 == \"admin\" { \"Access Granted\" } else { \"Denied\" }")

	// 6. Complex Chain
	s.mustSetCell("A13", "6. Pipeline")
	s.mustSetCell("B13", "= [\"apple\", \"banana\", \"cherry\"]")
	s.mustSetCell("C13", "= map(B13, toUpper)")
	s.mustSetCell("D13", "= filter(C13, (s) -> contains(s, \"A\"))") // APPLE, BANANA
	// 6. Complex Chain
	s.mustSetCell("A13", "6. Pipeline")
	s.mustSetCell("B13", "= [\"apple\", \"banana\", \"cherry\"]")
	s.mustSetCell("C13", "= map(B13, toUpper)")
	s.mustSetCell("D13", "= filter(C13, (s) -> contains(s, \"A\"))") // APPLE, BANANA
	s.mustSetCell("E13", "= D13.length")
}

func (s *Server) populateMatrix() {
	s.Sheet.Clear()

	s.mustSetCell("A1", "🚀 Reactive Matrix")

	// A2: Scalar Value
	s.mustSetCell("A2", "Scalar:")
	s.mustSetCell("B2", "2")
	s.mustSetCell("C2", "<- Change me!")

	// A3: Operation Function
	s.mustSetCell("A3", "Op:")
	s.mustSetCell("B3", "= (val, scalar) -> val * scalar")
	s.mustSetCell("C3", "<- Change logic! (e.g. Try: val + scalar)")

	// Header for Source Matrix (Row 5)
	s.mustSetCell("A5", "Source Matrix (10x10)")

	// Generate Source Matrix (A6:J15) -> Now 20x20 (A6:T25)
	// Values 1..400
	rows := 20
	cols := 20
	startRow := 6
	startCol := 1 // 'A'

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			val := r*cols + c + 1
			colName := getColName(startCol + c)
			cellID := fmt.Sprintf("%s%d", colName, startRow+r)
			s.mustSetCell(CellID(cellID), fmt.Sprintf("%d", val))
		}
	}

	// Header for Result Matrix (Row 5, Col W -> 23)
	// Source is cols 1-20 (A-T). Let's leave gap U,V. Start Result at W.
	resStartCol := 23 // 'W'
	resHeaderCol := getColName(resStartCol)
	s.mustSetCell(CellID(resHeaderCol+"5"), "Result Matrix (= Op(Source, Scalar))")

	// Generate Result Matrix (W6:AP25)
	// Formula: = B3(SourceCell, B2)

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			// Source Ref
			srcColName := getColName(startCol + c)
			srcCellID := fmt.Sprintf("%s%d", srcColName, startRow+r)

			// Result Ref
			resColName := getColName(resStartCol + c)
			resCellID := fmt.Sprintf("%s%d", resColName, startRow+r)

			// Formula: = B3(A6, B2)
			formula := fmt.Sprintf("= B3(%s, B2)", srcCellID)
			s.mustSetCell(CellID(resCellID), formula)
		}
	}

	// Aggregates (Row 27)
	sumRow := startRow + rows + 1
	s.mustSetCell(CellID(fmt.Sprintf("A%d", sumRow)), "Col Sums:")

	// Sum each Result Column
	for c := 0; c < cols; c++ {
		resColName := getColName(resStartCol + c)

		var additionChain string
		for r := 0; r < rows; r++ {
			cellID := fmt.Sprintf("%s%d", resColName, startRow+r)
			if r == 0 {
				additionChain = "= " + cellID
			} else {
				additionChain += " + " + cellID
			}
		}

		sumCellID := fmt.Sprintf("%s%d", resColName, sumRow)
		s.mustSetCell(CellID(sumCellID), additionChain)
	}

	// Grand Total
	// Place below Col Sums (Row 29)
	grandTotalRow := sumRow + 2
	grandTotalValCell := fmt.Sprintf("B%d", grandTotalRow)

	s.mustSetCell(CellID(fmt.Sprintf("A%d", grandTotalRow)), "Grand Total:")

	// Sum of Row 27 (Result Cols W..AP)
	var grandTotal string
	for c := 0; c < cols; c++ {
		// Fix: Use correct multi-char column logic
		// 'W' is index 22 (0-based) or 23 (1-based).
		// We have resStartCol = 23.
		// A=1, Z=26, AA=27.
		// Current logic: rune('A' + resStartCol + c - 1) breaks after Z.

		colIdx := resStartCol + c // 1-based index of column
		resColName := getColName(colIdx)

		sumCellID := fmt.Sprintf("%s%d", resColName, sumRow)
		if c == 0 {
			grandTotal = "= " + sumCellID
		} else {
			grandTotal += " + " + sumCellID
		}
	}
	s.mustSetCell(CellID(grandTotalValCell), grandTotal)
}

// Helper to convert 1-based index to column name (1->A, 26->Z, 27->AA)
func getColName(n int) string {
	name := ""
	for n > 0 {
		n--
		name = string(rune('A'+(n%26))) + name
		n /= 26
	}
	return name
}

func (s *Server) populateRanges() {
	s.Sheet.Clear()

	s.mustSetCell("A1", "🚀 Range Simulation")
	s.mustSetCell("B1", "(Manual Range Logic)")

	// Create a list of values in col A
	s.mustSetCell("A3", "Data:")
	s.mustSetCell("A4", "10")
	s.mustSetCell("A5", "20")
	s.mustSetCell("A6", "30")
	s.mustSetCell("A7", "40")
	s.mustSetCell("A8", "50")

	// Create a 'Range' object manually since we don't have native Range types yet
	// A range can be represented as a list of values? Or better, we build a list from the cells manually.

	s.mustSetCell("C3", "Range Object (List):")
	s.mustSetCell("C4", "= A4:A8")
	s.mustSetCell("D4", "<- Expanded Range!")

	// Operations on the "Range"
	s.mustSetCell("C6", "Sum:")
	s.mustSetCell("C7", "= reduce(C4, (acc, x) -> acc + x, 0)")

	s.mustSetCell("C9", "Average:")
	s.mustSetCell("C10", "= C7 / C4.length")

	s.mustSetCell("C12", "Identify Max:")
	// Reduce to find max
	s.mustSetCell("C13", "= reduce(C4, (max, x) -> if x > max { x } else { max }, 0)")

	s.mustSetCell("E3", "Dynamic Filter:")
	s.mustSetCell("E4", "Threshold:")
	s.mustSetCell("F4", "25")

	s.mustSetCell("E6", "Values > Threshold:")
	s.mustSetCell("E7", "= filter(C4, (x) -> x > F4)")
	s.mustSetCell("F7", "<- Updates automatically!")

	// Map operation: Scale values
	s.mustSetCell("E9", "Scaled (x2):")
	s.mustSetCell("E10", "= map(C4, (x) -> x * 2)")

	// Join string
	s.mustSetCell("E12", "Joined:")
	// Use reduce to join strings. Explicit type conversion isn't available yet as 'string()',
	// but we can rely on string concatenation behaviors if we ensure inputs are strings or handled.
	// Let's create a list of strings first.
	// IMPORTANT: Lists must reduce with correct accumulator type.
	// Initial accumulator is "" (string).
	// List is strings.
	// Op is string + string.
	// If the interpreter fails, it might be due to type checking on '+' or list element evaluation.
	// Let's ensure explicit valid list syntax.
	s.mustSetCell("E13", "= [\"a\", \"b\", \"c\"]") // Double quotes standard
	s.mustSetCell("E14", "= reduce(E13, (acc, x) -> acc + x, \"\")")
	// For numbers, we might need a custom join if 'string(int)' isn't there.

	// Dynamic List Construction
	s.mustSetCell("A17", "Dynamic List:")
	s.mustSetCell("B17", "Start:")
	s.mustSetCell("C17", "1")
	s.mustSetCell("B18", "Count:")
	s.mustSetCell("C18", "5")

	// Since we don't have a `range(start, count)` builtin function, we can't easily generate a list of arbitrary length dynamically *inside* a cell formula without recursion or a dedicated builtin.
	// BUT, we can simulate it if we had such a function.
	// For now, let's show how to build a list from a fixed set of cells that change values.

	s.mustSetCell("A20", "Generated:")
	// Let's make 5 cells that depend on Start
	s.mustSetCell("B20", "= C17")
	s.mustSetCell("B21", "= B20 + 1")
	s.mustSetCell("B22", "= B21 + 1")
	s.mustSetCell("B23", "= B22 + 1")
	s.mustSetCell("B24", "= B23 + 1")

	s.mustSetCell("C20", "= [B20, B21, B22, B23, B24]")
	s.mustSetCell("D20", "<- List [Start..Start+4]")
	s.mustSetCell("E20", "= reduce(C20, (a, b) -> a + b, 0)")
	s.mustSetCell("F20", "<- Sum of Dynamic List")

	s.mustSetCell("A1", "🚀 Range Simulation") // re-set for safety
}

func (s *Server) populateFactorial() {
	s.Sheet.Clear()

	s.mustSetCell("A1", "🚀 Recursive Factorial")

	// Define Factorial Function
	s.mustSetCell("A3", "Factorial Fn:")
	// Recursion in anonymous functions requires a way to reference itself.
	// Karl doesn't support named recursive let bindings easily inside a cell yet (unless we use Y-combinator or global define).
	// BUT, we can define it and assign it to a cell, then reference the CELL.
	// B3 = (n) -> if n <= 1 { 1 } else { n * B3(n - 1) }
	// PROBLEM: Cell references in Karl are evaluated to their *values*, not lazy references, usually.
	// If B3 depends on B3, we have a cycle. The dependency graph will block it.

	// ALTERNATIVE: Use the Y-Combinator for anonymous recursion!
	// Y = (f) -> ((x) -> x(x))((x) -> f((y) -> x(x)(y)))
	// Factorial = Y((fact) -> (n) -> if n <= 1 { 1 } else { n * fact(n - 1) })

	s.mustSetCell("B3", "=(f)->((x)->f((v)->x(x)(v)))((x)->f((v)->x(x)(v)))")
	s.mustSetCell("C3", "<- Y-Combinator (Z-Comb for strict)")

	s.mustSetCell("A5", "Fact Gen:")
	s.mustSetCell("B5", "=(fact) -> (n) -> if n <= 1 { 1 } else { n * fact(n - 1) }")

	s.mustSetCell("A7", "Factorial:")
	s.mustSetCell("B7", "= B3(B5)") // The recursive function
	s.mustSetCell("C7", "<- Ready to use!")

	// Test it
	s.mustSetCell("A9", "Input:")
	s.mustSetCell("B9", "5")

	s.mustSetCell("A10", "Result:")
	s.mustSetCell("B10", "= B7(B9)")

	// Visual sequence (Now deeper)
	s.mustSetCell("D1", "Sequence:")
	// Show inputs in C and results in D
	for i := 1; i <= 20; i++ { // Extended to 20
		row := i + 2
		// Input
		inputID := fmt.Sprintf("C%d", row)
		s.mustSetCell(CellID(inputID), fmt.Sprintf("%d", i))

		// Calc logic: Only show if i <= B9 (input value)
		// We use `if` inside the Karl formula.
		// D_i = if C_i <= B9 { B7(C_i) } else { "" }

		resultID := fmt.Sprintf("D%d", row)
		// We also want to hide the input "C" column numbers if > B9?
		// Or maybe keep inputs visible but results hidden?
		// User asked: "I want B3:B12 to be those 10 cells" (implied he wants the list to grow/shrink).
		// Let's make the Result cell conditional.

		formula := fmt.Sprintf("= if %s <= B9 { B7(%s) } else { \"\" }", inputID, inputID)
		s.mustSetCell(CellID(resultID), formula)
	}
}

type UpdateRequest struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Value   string `json:"value"`
	Example string `json:"example,omitempty"`
}

type UpdateResponse struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Value   string `json:"value"`
	Display string `json:"display"`
	Error   string `json:"error,omitempty"`
}

func (s *Server) sendInitialState(conn *websocket.Conn) {
	s.Sheet.mu.RLock()
	defer s.Sheet.mu.RUnlock()

	for _, cell := range s.Sheet.Cells {
		resp := s.createUpdateResponse(cell)
		if err := conn.WriteJSON(resp); err != nil {
			log.Printf("initial state write failed: %v", err)
			return
		}
	}
}

func (s *Server) handleUpdate(req UpdateRequest) {
	id := CellID(req.ID)
	// Update sheet
	if err := s.Sheet.SetCell(id, req.Value); err != nil {
		log.Printf("Error setting cell %s: %v", id, err)
	}

	// Broadcast updates for this cell and its dependents
	// We need to know which cells changed.
	// For now, let's just broadcast the target cell and let the sheet logic tell us what else changed?
	// Our Sheet implementation doesn't return list of changed cells yet.
	// Let's modify SetCell to return changed cells or just broadcast everything for MVP?
	// Broadcasting everything is inefficient.
	// Let's just broadcast the modified cell for now, and relying on `SetCell` to have updated the values.
	// BUT `SetCell` updates dependents too. We need to find all cells that were re-evaluated.

	// IMPROVEMENT: We should inspect the sheet to find updated cells.
	// Or simplistic approach: Iterate all cells and send updates (inefficient but works).
	// Better: `SetCell` should return affected cells. I will update `SetCell` signature later.
	// For MVP, I'll just broadcast the requested cell AND all its dependents recursively.

	// Actually, let's just traverse the graph from the updated cell and broadcast all reachable dependents.

	affected := make(map[CellID]bool)
	s.collectAffected(id, affected)

	s.broadcastUpdates(affected)
}

func (s *Server) collectAffected(id CellID, affected map[CellID]bool) {
	if affected[id] {
		return
	}
	affected[id] = true

	cell := s.Sheet.GetCell(id) // Thread-safe
	s.Sheet.mu.RLock()
	dependents := make([]CellID, len(cell.Dependents))
	copy(dependents, cell.Dependents)
	s.Sheet.mu.RUnlock()

	for _, dep := range dependents {
		s.collectAffected(dep, affected)
	}
}

func (s *Server) broadcastUpdates(affected map[CellID]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id := range affected {
		cell := s.Sheet.GetCell(id)
		resp := s.createUpdateResponse(cell)

		for client := range s.clients {
			if err := client.WriteJSON(resp); err != nil {
				log.Printf("update write failed: %v", err)
				_ = client.Close()
				delete(s.clients, client)
			}
		}
	}
}

func (s *Server) createUpdateResponse(cell *Cell) UpdateResponse {
	// Need to format the value nicely
	valStr := ""
	if cell.Error != nil {
		valStr = "#ERROR"
	} else if cell.Value != nil {
		if val, ok := cell.Value.(interpreter.Value); ok {
			valStr = val.Inspect()
			// Unquote string literal results for cleaner display
			if val.Type() == interpreter.STRING {
				if len(valStr) >= 2 && valStr[0] == '"' && valStr[len(valStr)-1] == '"' {
					valStr = valStr[1 : len(valStr)-1]
				}
			}
		} else {
			// Fallback (should ideally be unreachable if everything is Value)
			valStr = fmt.Sprintf("%v", cell.Value)
		}
	}

	return UpdateResponse{
		Type:    "cell_updated",
		ID:      string(cell.ID),
		Value:   cell.RawValue,
		Display: valStr,
		Error: func() string {
			if cell.Error != nil {
				return cell.Error.Error()
			}
			return ""
		}(),
	}
}

// Start starts the HTTP server on the given address.
// Start starts the HTTP server on the given address.
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// Serve static files
	dir := "assets/spreadsheet"
	// Check if dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("Warning: Static directory %s not found. Current dir: %s", dir, func() string { d, _ := os.Getwd(); return d }())
	} else {
		log.Printf("Serving static files from %s", dir)
	}

	fs := http.FileServer(http.Dir(dir))
	mux.Handle("/", fs)

	mux.HandleFunc("/ws", s.HandleWebSocket)

	log.Printf("Starting spreadsheet server at http://%s", addr)
	return http.ListenAndServe(addr, mux)
}
