package main

import (
	"strings"
	"testing"

	"karl/interpreter"
)

func TestParseRunArgsProgramArgsSeparator(t *testing.T) {
	policy, positional, programArgs, help, err := parseRunArgs([]string{"app.k", "--", "a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if help {
		t.Fatalf("expected help=false")
	}
	if policy != interpreter.TaskFailurePolicyFailFast {
		t.Fatalf("expected default policy %q, got %q", interpreter.TaskFailurePolicyFailFast, policy)
	}
	if len(positional) != 1 || positional[0] != "app.k" {
		t.Fatalf("unexpected positional: %#v", positional)
	}
	if len(programArgs) != 2 || programArgs[0] != "a" || programArgs[1] != "b" {
		t.Fatalf("unexpected programArgs: %#v", programArgs)
	}
}

func TestParseRunArgsProgramArgsCanLookLikeFlags(t *testing.T) {
	policy, positional, programArgs, help, err := parseRunArgs([]string{
		"--task-failure-policy=defer",
		"app.k",
		"--",
		"-x",
		"--y",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if help {
		t.Fatalf("expected help=false")
	}
	if policy != interpreter.TaskFailurePolicyDefer {
		t.Fatalf("expected policy %q, got %q", interpreter.TaskFailurePolicyDefer, policy)
	}
	if len(positional) != 1 || positional[0] != "app.k" {
		t.Fatalf("unexpected positional: %#v", positional)
	}
	if len(programArgs) != 2 || programArgs[0] != "-x" || programArgs[1] != "--y" {
		t.Fatalf("unexpected programArgs: %#v", programArgs)
	}
}

func TestParseRunArgsRejectsProgramArgsWithoutSeparator(t *testing.T) {
	_, _, _, _, err := parseRunArgs([]string{"app.k", "a", "b"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "program args must follow `--`") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseRunArgsStdinWithProgramArgs(t *testing.T) {
	policy, positional, programArgs, help, err := parseRunArgs([]string{"-", "--", "foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if help {
		t.Fatalf("expected help=false")
	}
	if policy != interpreter.TaskFailurePolicyFailFast {
		t.Fatalf("expected default policy %q, got %q", interpreter.TaskFailurePolicyFailFast, policy)
	}
	if len(positional) != 1 || positional[0] != "-" {
		t.Fatalf("unexpected positional: %#v", positional)
	}
	if len(programArgs) != 1 || programArgs[0] != "foo" {
		t.Fatalf("unexpected programArgs: %#v", programArgs)
	}
}

func TestParseDebugArgsUsesRunContract(t *testing.T) {
	policy, positional, programArgs, help, err := parseDebugArgs([]string{"app.k", "--", "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if help {
		t.Fatalf("expected help=false")
	}
	if policy != interpreter.TaskFailurePolicyFailFast {
		t.Fatalf("expected policy %q, got %q", interpreter.TaskFailurePolicyFailFast, policy)
	}
	if len(positional) != 1 || positional[0] != "app.k" {
		t.Fatalf("unexpected positional: %#v", positional)
	}
	if len(programArgs) != 1 || programArgs[0] != "x" {
		t.Fatalf("unexpected programArgs: %#v", programArgs)
	}
}

func TestParseBreakpointSpecLineOnly(t *testing.T) {
	file, line, err := parseBreakpointSpec("42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file != "" || line != 42 {
		t.Fatalf("unexpected breakpoint: file=%q line=%d", file, line)
	}
}

func TestParseBreakpointSpecFileAndLine(t *testing.T) {
	file, line, err := parseBreakpointSpec("app.k:12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file != "app.k" || line != 12 {
		t.Fatalf("unexpected breakpoint: file=%q line=%d", file, line)
	}
}

func TestParseBreakpointSpecInvalid(t *testing.T) {
	_, _, err := parseBreakpointSpec("bad")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestEvalDebugExpressionUsesCurrentEnv(t *testing.T) {
	env := interpreter.NewBaseEnvironment()
	env.Define("x", &interpreter.Integer{Value: 41})

	val, err := evalDebugExpression("x + 1", env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	i, ok := val.(*interpreter.Integer)
	if !ok {
		t.Fatalf("expected Integer, got %T", val)
	}
	if i.Value != 42 {
		t.Fatalf("expected 42, got %d", i.Value)
	}
}

func TestEvalDebugExpressionRejectsStatement(t *testing.T) {
	env := interpreter.NewBaseEnvironment()
	_, err := evalDebugExpression("let x = 1", env)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "expression") {
		t.Fatalf("unexpected error: %v", err)
	}
}
