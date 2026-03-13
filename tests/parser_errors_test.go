package tests

import (
	"karl/lexer"
	"karl/parser"
	"strings"
	"testing"
)

func TestParserErrors(t *testing.T) {
	cases := []struct {
		name         string
		input        string
		errorContain string
	}{
		{
			name:  "float_range_not_allowed",
			input: "1.0..2.0",
		},
		{
			name:  "spawn_requires_call",
			input: "& foo",
		},
		{
			name:  "spawn_keyword_requires_call",
			input: "spawn foo",
		},
		{
			name:         "spawn_operator_for_expr_rejected",
			input:        "& for i < 3 with i = 0 { i++ } then i",
			errorContain: "spawn target must be a call expression or stream pipeline",
		},
		{
			name:         "spawn_keyword_for_expr_rejected",
			input:        "spawn (for i < 3 with i = 0 { i++ } then i)",
			errorContain: "spawn target must be a call expression or stream pipeline",
		},
		{
			name:  "object_shorthand_requires_trailing_comma",
			input: "let o = { x, y }",
		},
		{
			name:         "pipe_without_left_operand",
			input:        "wait | { fast(), slow() }",
			errorContain: "reserved for stream pipelines",
		},
		{
			name:         "spawn_proc_is_rejected",
			input:        "& proc({ command: \"echo\", args: [\"hi\"], })",
			errorContain: "already returns <process>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.input))
			_ = p.ParseProgram()
			errors := p.Errors()
			if len(errors) == 0 {
				t.Fatalf("expected parse errors")
			}
			if tc.errorContain != "" {
				for _, err := range errors {
					if strings.Contains(err, tc.errorContain) {
						return
					}
				}
				t.Fatalf("expected parse error containing %q, got %v", tc.errorContain, errors)
			}
		})
	}
}
