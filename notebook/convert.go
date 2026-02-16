package notebook

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// JupyterNotebook represents a .ipynb file structure
type JupyterNotebook struct {
	Cells []JupyterCell `json:"cells"`
}

// JupyterCell represents a cell in a .ipynb file
type JupyterCell struct {
	CellType string      `json:"cell_type"`
	Source   interface{} `json:"source"`
}

// ConvertCommand handles the convert subcommand
func ConvertCommand(args []string) int {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: karl notebook convert <input.ipynb> <output.knb>\n")
		return 2
	}

	inputFile := args[0]
	outputFile := args[1]

	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		return 1
	}

	var jnb JupyterNotebook
	if err := json.Unmarshal(data, &jnb); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing Jupyter notebook: %v\n", err)
		return 1
	}

	knb := &Notebook{
		Title:       "Imported Notebook",
		Description: "Converted from " + inputFile,
		Version:     "1.0",
		Cells:       make([]Cell, 0),
	}

	for _, jCell := range jnb.Cells {
		var content string
		
		// Jupyter source can be string or []string
		switch src := jCell.Source.(type) {
		case string:
			content = src
		case []interface{}:
			var lines []string
			for _, line := range src {
				if s, ok := line.(string); ok {
					lines = append(lines, s)
				}
			}
			content = strings.Join(lines, "")
		default:
			continue
		}

		var kCell Cell
		if jCell.CellType == "code" {
			kCell = Cell{
				Type:   CodeCell,
				Source: content,
			}
		} else if jCell.CellType == "markdown" {
			kCell = Cell{
				Type:   MarkdownCell,
				Source: content,
			}
		} else {
			continue // Skip unknown types
		}
		
		knb.Cells = append(knb.Cells, kCell)
	}

	outData, err := json.MarshalIndent(knb, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Karl notebook: %v\n", err)
		return 1
	}

	if err := os.WriteFile(outputFile, outData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		return 1
	}

	fmt.Printf("Successfully converted %s to %s\n", inputFile, outputFile)
	return 0
}
