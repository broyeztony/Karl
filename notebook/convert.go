package notebook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// JupyterNotebook represents a .ipynb file structure.
type JupyterNotebook struct {
	Cells         []JupyterCell          `json:"cells"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	NBFormat      int                    `json:"nbformat"`
	NBFormatMinor int                    `json:"nbformat_minor"`
}

// JupyterCell represents a cell in a .ipynb file.
type JupyterCell struct {
	ID             string                 `json:"id,omitempty"`
	CellType       string                 `json:"cell_type"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Source         interface{}            `json:"source"`
	ExecutionCount interface{}            `json:"execution_count,omitempty"`
	Outputs        []interface{}          `json:"outputs,omitempty"`
}

func formatFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".ipynb":
		return "ipynb"
	case ".knb":
		return "knb"
	default:
		return ""
	}
}

func loadIPYNB(filename string) (*Notebook, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read notebook: %w", err)
	}

	var jnb JupyterNotebook
	if err := json.Unmarshal(data, &jnb); err != nil {
		return nil, fmt.Errorf("failed to parse Jupyter notebook: %w", err)
	}

	title := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	nb := &Notebook{
		Title:       title,
		Description: "",
		Version:     "1.0",
		Cells:       make([]Cell, 0, len(jnb.Cells)),
		Metadata:    make(map[string]interface{}),
	}

	for _, jCell := range jnb.Cells {
		source, ok := jupyterSourceToString(jCell.Source)
		if !ok {
			continue
		}

		switch jCell.CellType {
		case "code":
			nb.Cells = append(nb.Cells, Cell{Type: CodeCell, Source: source, Metadata: map[string]interface{}{}})
		case "markdown":
			nb.Cells = append(nb.Cells, Cell{Type: MarkdownCell, Source: source, Metadata: map[string]interface{}{}})
		}
	}

	return nb, nil
}

func saveIPYNB(nb *Notebook, filename string) error {
	cells := make([]map[string]interface{}, 0, len(nb.Cells))
	for i, cell := range nb.Cells {
		cellID := fmt.Sprintf("karl-cell-%d", i+1)
		switch cell.Type {
		case CodeCell:
			cells = append(cells, map[string]interface{}{
				"id":              cellID,
				"cell_type":       "code",
				"metadata":        map[string]interface{}{},
				"source":          splitLinesForJupyter(cell.Source),
				"execution_count": nil,
				"outputs":         []interface{}{},
			})
		case MarkdownCell:
			cells = append(cells, map[string]interface{}{
				"id":        cellID,
				"cell_type": "markdown",
				"metadata":  map[string]interface{}{},
				"source":    splitLinesForJupyter(cell.Source),
			})
		}
	}

	jnb := map[string]interface{}{
		"cells": cells,
		"metadata": map[string]interface{}{
			"kernelspec": map[string]interface{}{
				"display_name": "Karl",
				"language":     "karl",
				"name":         "karl",
			},
			"language_info": map[string]interface{}{
				"name":           "karl",
				"file_extension": ".k",
			},
		},
		"nbformat":       4,
		"nbformat_minor": 5,
	}

	data, err := json.MarshalIndent(jnb, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Jupyter notebook: %w", err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write notebook: %w", err)
	}
	return nil
}

func jupyterSourceToString(src interface{}) (string, bool) {
	switch v := src.(type) {
	case string:
		return v, true
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, part := range v {
			s, ok := part.(string)
			if !ok {
				continue
			}
			parts = append(parts, s)
		}
		return strings.Join(parts, ""), true
	default:
		return "", false
	}
}

func splitLinesForJupyter(source string) []string {
	if source == "" {
		return []string{}
	}
	lines := strings.SplitAfter(source, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// ConvertCommand handles the convert subcommand.
func ConvertCommand(args []string) int {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: karl notebook convert <input.{ipynb|knb}> <output.{ipynb|knb}>\n")
		return 2
	}

	inputFile := args[0]
	outputFile := args[1]
	inFmt := formatFromFilename(inputFile)
	outFmt := formatFromFilename(outputFile)

	if inFmt == "" || outFmt == "" {
		fmt.Fprintf(os.Stderr, "unsupported notebook format (expected .ipynb or .knb)\n")
		return 2
	}

	var nb *Notebook
	var err error
	if inFmt == "knb" {
		nb, err = LoadNotebook(inputFile)
	} else {
		nb, err = loadIPYNB(inputFile)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading input notebook: %v\n", err)
		return 1
	}

	if outFmt == "knb" {
		err = nb.SaveNotebook(outputFile)
	} else {
		err = saveIPYNB(nb, outputFile)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving output notebook: %v\n", err)
		return 1
	}

	fmt.Printf("Successfully converted %s (%s) -> %s (%s)\n", inputFile, inFmt, outputFile, outFmt)
	return 0
}
