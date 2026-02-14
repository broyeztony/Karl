package interpreter

import (
	"fmt"
	"strings"

	"karl/token"
)

type RuntimeError struct {
	Message string
	Token   *token.Token
}

func (e *RuntimeError) Error() string {
	return e.Message
}

type RecoverableError struct {
	Message string
	Kind    string
	Token   *token.Token
}

func (e *RecoverableError) Error() string {
	return e.Message
}

func FormatRuntimeError(err error, source string, filename string) string {
	switch e := err.(type) {
	case *RuntimeError:
		return formatRuntimeError(e.Message, e.Token, source, filename)
	case *RecoverableError:
		return formatRuntimeError(e.Message, e.Token, source, filename)
	default:
		return err.Error()
	}
}

func formatRuntimeError(message string, tok *token.Token, source string, filename string) string {
	if tok == nil || tok.Line == 0 || source == "" {
		return "runtime error: " + message
	}
	lines := strings.Split(source, "\n")
	line := tok.Line
	col := tok.Column
	if line < 1 || line > len(lines) {
		return "runtime error: " + message
	}
	lineText := strings.TrimRight(lines[line-1], "\r")
	if col < 1 {
		col = 1
	}
	if col > len(lineText)+1 {
		col = len(lineText) + 1
	}
	caret := strings.Repeat(" ", col-1) + "^"
	location := fmt.Sprintf("%d:%d", line, tok.Column)
	if filename != "" {
		location = fmt.Sprintf("%s:%s", filename, location)
	}
	return fmt.Sprintf(
		"runtime error: %s\n  at %s\n  %d | %s\n    | %s",
		message,
		location,
		line,
		lineText,
		caret,
	)
}

type ExitError struct {
	Message string
}

func (e *ExitError) Error() string {
	if e.Message == "" {
		return "exit"
	}
	return fmt.Sprintf("exit: %s", e.Message)
}

// UnhandledTaskError is returned by the CLI runner when one or more tasks failed
// and nobody awaited/handled them.
//
// The Messages are already formatted (they may reference different source files),
// so callers should print Error() directly (and not re-wrap via FormatRuntimeError).
type UnhandledTaskError struct {
	Messages []string
}

func (e *UnhandledTaskError) Error() string {
	if e == nil || len(e.Messages) == 0 {
		return "unhandled task failure"
	}
	return "unhandled task failures:\n\n" + strings.Join(e.Messages, "\n\n")
}
