package notebook

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"karl/lexer"
	"karl/parser"
	"karl/repl"
)

// RunInteractive executes the notebook in the specified mode.
func (r *Runner) RunInteractive(nb *Notebook, step bool, replMode bool, filename string) error {
	scanner := bufio.NewScanner(os.Stdin)

	// Phase 1: Execute existing cells
	if step {
		// Step-by-Step Execution
		fmt.Println("Starting Step-by-Step Execution...")
		fmt.Println("Commands: [Enter] Run, [s] Skip, [q] Quit")
		
		for i, cell := range nb.Cells {
			if cell.Type != CodeCell {
				continue
			}

			fmt.Printf("\n--- Cell %d ---\n", i+1)
			fmt.Println(cell.Source)
			fmt.Print("\n(Run/s/q) > ")
			
			if !scanner.Scan() {
				break
			}
			input := strings.TrimSpace(strings.ToLower(scanner.Text()))
			
			if input == "q" {
				fmt.Println("Aborting.")
				return nil
			}
			if input == "s" {
				fmt.Println("Skipping.")
				continue
			}
			
			// Execute
			output, err := r.ExecuteCell(i, cell)
			if err != nil {
				return err
			}
			
			if output.Error != nil {
				fmt.Printf("Error: %s\n", output.Error.Message)
			} else if output.Value != "" {
				fmt.Printf("Output: %s\n", output.Value)
			}
		}
	} else {
		// Batch Execution (silent/fast-forward) if entering REPL
		// If both false, main.go handles standard Batch output.
		// If replMode is true, we need to execute state first.
		if replMode {
			fmt.Println("Executing existing cells...")
			outputs, err := r.ExecuteNotebook(nb)
			
			for _, output := range outputs {
				fmt.Printf("\n--- Cell %d ---\n", output.CellIndex+1)
				// We need the source. accessing nb.Cells
				if output.CellIndex < len(nb.Cells) {
					fmt.Println(nb.Cells[output.CellIndex].Source)
				}
				
				if output.Error != nil {
					fmt.Printf("Error: %s\n", output.Error.Message)
				} else if output.Value != "" {
					fmt.Printf("Output: %s\n", output.Value)
				}
			}

			if err != nil {
				fmt.Printf("Error executing notebook: %v\n", err)
			}
		}
	}
	
	// Phase 2: REPL Mode
	if replMode {
		fmt.Println("\n--- Notebook REPL Mode ---")
		fmt.Println("Type code and press Enter. New cells will be appended to the notebook.")
		fmt.Println("Commands: :quit")
		
		var inputBuffer strings.Builder
		multiline := false

		for {
			prompt := "notebook> "
			if multiline {
				prompt = "...       "
			}
			fmt.Print(prompt)
			
			if !scanner.Scan() {
				break
			}
			line := scanner.Text()
			
			if !multiline && (line == ":quit" || line == ":q") {
				break
			}
			
			if inputBuffer.Len() > 0 {
				inputBuffer.WriteString("\n")
			}
			inputBuffer.WriteString(line)
			
			input := inputBuffer.String()
			
			l := lexer.New(input)
			p := parser.New(l)
			p.ParseProgram()
			errs := p.ErrorsDetailed()
			
			if repl.IsIncompleteInput(input, errs) {
				multiline = true
				continue
			}
			
			if len(errs) > 0 {
				fmt.Printf("Parse error: %v\n", errs)
				inputBuffer.Reset()
				multiline = false
				continue
			}
			
			// Execute
			multiline = false
			inputBuffer.Reset()
			
			// Create a temporary cell to execute
			tempCell := Cell{
				Type:   CodeCell,
				Source: input,
			}
			
			// Use next index
			idx := len(nb.Cells)
			output, err := r.ExecuteCell(idx, tempCell)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			
			if output.Error != nil {
				fmt.Printf("Runtime Error: %s\n", output.Error.Message)
			} else {
				if output.Value != "" {
					fmt.Println(output.Value)
				}
				
				// Append and Save
				nb.AddCodeCell(input, "")
				if err := nb.SaveNotebook(filename); err != nil {
					fmt.Printf("Warning: Failed to save notebook: %v\n", err)
				} else {
					// fmt.Println("Wait, saved.") // Removed annoying log
				}
			}
		}
	}

	return nil
}
