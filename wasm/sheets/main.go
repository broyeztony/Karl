//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"karl/spreadsheet"
	"syscall/js"
)

type sheetCommand struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Value   string `json:"value,omitempty"`
	Example string `json:"example,omitempty"`
}

type sheetCommandResult struct {
	Reset    bool                         `json:"reset,omitempty"`
	Messages []spreadsheet.UpdateResponse `json:"messages,omitempty"`
	Error    string                       `json:"error,omitempty"`
}

var runtime *spreadsheet.Runtime

func main() {
	runtime = spreadsheet.NewRuntime()
	js.Global().Set("runKarlSheets", js.FuncOf(runKarlSheets))
	fmt.Println("Karl Sheets WASM Runtime initialized.")
	select {}
}

func runKarlSheets(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return marshalResult(sheetCommandResult{Error: "runKarlSheets expects 1 argument (json command)"})
	}

	var cmd sheetCommand
	if err := json.Unmarshal([]byte(args[0].String()), &cmd); err != nil {
		return marshalResult(sheetCommandResult{Error: "invalid command: " + err.Error()})
	}

	switch cmd.Type {
	case "init":
		return marshalResult(sheetCommandResult{Messages: runtime.Snapshot()})
	case "update_cell":
		if cmd.ID == "" {
			return marshalResult(sheetCommandResult{Error: "update_cell expects id"})
		}
		updates, err := runtime.UpdateCell(cmd.ID, cmd.Value)
		if err != nil {
			return marshalResult(sheetCommandResult{Error: err.Error()})
		}
		return marshalResult(sheetCommandResult{Messages: updates})
	case "clear":
		runtime.Clear()
		return marshalResult(sheetCommandResult{Reset: true})
	case "load_example":
		runtime.LoadExample(cmd.Example)
		return marshalResult(sheetCommandResult{
			Reset:    true,
			Messages: runtime.Snapshot(),
		})
	default:
		return marshalResult(sheetCommandResult{Error: "unknown command: " + cmd.Type})
	}
}

func marshalResult(out sheetCommandResult) string {
	data, err := json.Marshal(out)
	if err != nil {
		return `{"error":"marshal error"}`
	}
	return string(data)
}
